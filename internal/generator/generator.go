package generator

import (
	"math/rand"
	"sort"

	"dp-calc/types"
)

func GenerateTasksWithPolishing(req types.GenerateRequest) []types.Job {
	// 1) Случайно генерируем setupTimes и processTimes
	setupTimes := make([]int, req.NumTypes)
	processTimes := make([]int, req.NumTypes)
	for i := 0; i < req.NumTypes; i++ {
		setupTimes[i] = rand.Intn(req.MaxSetup-req.MinSetup+1) + req.MinSetup
		processTimes[i] = rand.Intn(req.MaxProcess-req.MinProcess+1) + req.MinProcess
	}

	// 2) Собираем jobs без дедлайнов
	N := req.NumTypes * req.PiecesPerType
	allJobs := make([]types.Job, 0, N)
	for t := 1; t <= req.NumTypes; t++ {
		for inst := 1; inst <= req.PiecesPerType; inst++ {
			allJobs = append(
				allJobs, types.Job{
					Type:     t,
					Instance: inst,
					Setup:    setupTimes[t-1],
					Process:  processTimes[t-1],
					Deadline: 0, // заполним ниже
				},
			)
		}
	}

	// 3) Перемешиваем jobs (случайная последовательность)
	seq := rand.Perm(N)

	// 4) Вычисляем точные completion-times для seq (с учётом setup, process)
	completion := make([]int, N)
	currTime := 0
	for i, idx := range seq {
		j := allJobs[idx]
		if i == 0 {
			currTime += j.Setup + j.Process
		} else {
			prev := allJobs[seq[i-1]]
			if prev.Type == j.Type {
				currTime += j.Process
			} else {
				currTime += j.Setup + j.Process
			}
		}
		completion[i] = currTime
	}

	// 5) Назначаем дедлайны = completion + рандом [0..bigSlack] + небольшое смещение по типу [0..typeSlack]
	//    Сохраняем строго неубывающее свойство дедлайнов (чтобы EDD-проверка гарантированно прошла).
	lastDeadline := 0
	for i, idx := range seq {
		base := completion[i]

		// если base <= lastDeadline, двигаем base вперёд, чтобы дедлайны строго росли
		if base <= lastDeadline {
			base = lastDeadline + 1
		}

		// добавляем случайный запас 0..bigSlack
		sl := rand.Intn(req.BigSlack + 1)
		// добавляем «случайный сдвиг по типу» (0..typeSlack),
		// чтобы дедлайны одного типа немного «рассыпались»
		typeOffset := rand.Intn(5 + 1)

		dead := base + sl + typeOffset
		// dead := base + sl
		allJobs[idx].Deadline = dead
		lastDeadline = dead
	}

	// 6) СОРТИРУЕМ DEADLINES ВНУТРИ КАЖДОГО TYPE
	// Сгруппируем по типу указатели на allJobs
	byType := make(map[int][]*types.Job, req.NumTypes)
	for i := range allJobs {
		j := &allJobs[i]
		byType[j.Type] = append(byType[j.Type], j)
	}
	// Для каждого типа — сортируем срез по Deadline и переприсваиваем Instance
	for t := 1; t <= req.NumTypes; t++ {
		slice := byType[t]
		sort.Slice(
			slice, func(a, b int) bool {
				return slice[a].Deadline < slice[b].Deadline
			},
		)
		for rank, pj := range slice {
			pj.Instance = rank + 1
		}
	}

	return allJobs
}

// EDDCheck проверяет, что задачи в jobs могут быть выполнены в порядке EDD (Earliest Due Date).
func EDDCheck(jobs []types.Job) bool {
	sorted := make([]types.Job, len(jobs))
	copy(sorted, jobs)
	sort.Slice(
		sorted, func(i, j int) bool {
			return sorted[i].Deadline < sorted[j].Deadline
		},
	)
	ct := 0
	for _, j := range sorted {
		ct += j.Setup + j.Process
		if ct > j.Deadline {
			return false
		}
	}
	return true
}
