package handles

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"dp-calc/internal/dp"
	"dp-calc/internal/generator"
	"dp-calc/internal/gms"
	"dp-calc/types"
)

func HandleCompute(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Неверные параметры", http.StatusBadRequest)
		return
	}
	// Базовые T и P
	T, _ := strconv.Atoi(r.FormValue("numTypes"))
	P, _ := strconv.Atoi(r.FormValue("piecesPerType"))
	if T < 1 || P < 1 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Собираем jobs
	var jobs []types.Job
	if r.FormValue("generate") == "1" {
		// 2) Читаем диапазоны генератора
		minS, _ := strconv.Atoi(r.FormValue("minSetup"))
		maxS, _ := strconv.Atoi(r.FormValue("maxSetup"))
		minPr, _ := strconv.Atoi(r.FormValue("minProcess"))
		maxPr, _ := strconv.Atoi(r.FormValue("maxProcess"))

		// 3) Генерируем до тех пор, пока не пройдёт EDD-проверка
		for {
			candidate := generator.GenerateTasksWithPolishing(
				types.GenerateRequest{
					NumTypes:      T,
					PiecesPerType: P,
					MinSetup:      minS,
					MaxSetup:      maxS,
					MinProcess:    minPr,
					MaxProcess:    maxPr,
				},
			)
			if generator.EDDCheck(candidate) {
				jobs = candidate
				break
			}
		}
	} else {
		// Ручной ввод техпараметров
		setup := make([]int, T)
		process := make([]int, T)
		deadlines := make([][]int, T)
		for i := 0; i < T; i++ {
			setup[i], _ = strconv.Atoi(r.FormValue(fmt.Sprintf("setup%d", i)))
			process[i], _ = strconv.Atoi(r.FormValue(fmt.Sprintf("process%d", i)))
			deadlines[i] = make([]int, P)
			for j := 0; j < P; j++ {
				deadlines[i][j], _ = strconv.Atoi(
					r.FormValue(fmt.Sprintf("deadline%d_%d", i, j)),
				)
			}
		}
		for t := 1; t <= T; t++ {
			for inst := 1; inst <= P; inst++ {
				jobs = append(
					jobs, types.Job{
						Type:     t,
						Instance: inst,
						Setup:    setup[t-1],
						Process:  process[t-1],
						Deadline: deadlines[t-1][inst-1],
					},
				)
			}
		}
	}

	// Сохраняем текст параметров для отображения
	var table [][]string
	// Заголовок
	header := []string{"Type", "Setup", "Process"}
	for i := 1; i <= P; i++ {
		header = append(header, fmt.Sprintf("Deadline%d", i))
	}
	table = append(table, header)

	// Данные
	typeInfo := make(
		map[int]struct {
			setup     int
			process   int
			deadlines []int
		},
	)

	for _, j := range jobs {
		info := typeInfo[j.Type]
		if info.deadlines == nil {
			info.deadlines = make([]int, P)
			info.setup = j.Setup
			info.process = j.Process
		}
		info.deadlines[j.Instance-1] = j.Deadline
		typeInfo[j.Type] = info
	}

	for t := 1; t <= T; t++ {
		info := typeInfo[t]
		row := []string{
			fmt.Sprintf("%d", t),
			fmt.Sprintf("%d", info.setup),
			fmt.Sprintf("%d", info.process),
		}
		for _, d := range info.deadlines {
			row = append(row, fmt.Sprintf("%d", d))
		}
		table = append(table, row)
	}
	lastParamsTable = table

	// Запуск DP
	start := time.Now()
	dpFlat, base := dp.ComputeDP(jobs, T, P)
	elapsed := time.Since(start).Seconds()

	// Поиск оптимума
	fullCounts := make([]int, T)
	for i := range fullCounts {
		fullCounts[i] = P
	}
	key := dp.EncodeCountsToCode(fullCounts, base)
	minVal, bestLast := types.INF, 0
	for last := 1; last <= T; last++ {
		v := dpFlat[key*(T+1)+last]
		if v < minVal {
			minVal, bestLast = v, last
		}
	}
	if minVal == types.INF {
		types.TPLError.Execute(w, nil)
		return
	}

	// Сохраняем GMS
	lastGMS = gms.BuildGMS(jobs, T, P)

	// Реконструируем расписание
	schedule := dp.ReconstructSchedule(dpFlat, jobs, T, P, base, bestLast)
	steps, ticks := dp.BuildStepsFromSchedule(schedule)

	data := struct {
		T, P        int
		MinVal      int
		BestLast    int
		ParamsTable [][]string
		Steps       []types.Step
		Ticks       []int
		DPTime      float64
	}{
		T: T, P: P, MinVal: ticks[len(ticks)-1], BestLast: bestLast, ParamsTable: lastParamsTable,
		Steps: steps, Ticks: ticks, DPTime: elapsed,
	}
	types.TPLResult.Execute(w, data)
}

func HandleIndex(w http.ResponseWriter, _ *http.Request) {
	types.TPLIndex.Execute(w, nil)
}

func HandleParams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	T, _ := strconv.Atoi(r.FormValue("numTypes"))
	P, _ := strconv.Atoi(r.FormValue("piecesPerType"))
	data := struct {
		T      int
		P      int
		RangeT []int
		RangeP []int
	}{T, P, make([]int, T), make([]int, P)}
	for i := range data.RangeT {
		data.RangeT[i] = i
	}
	for j := range data.RangeP {
		data.RangeP[j] = j
	}
	types.TPLParams.Execute(w, data)
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=input.gms")
	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, bytes.NewReader(lastGMS))
}

var lastParamsTable [][]string
var lastGMS []byte
