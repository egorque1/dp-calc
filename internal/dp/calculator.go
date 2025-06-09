package dp

import (
	"runtime"
	"sync"

	"dp-calc/types"
)

// ComputeDP выполняет динамическое программирование для задачи планирования работ.
func ComputeDP(jobs []types.Job, numTypes, piecesPerType int) ([]int, int) {
	base := piecesPerType + 1
	numStates := 1
	for i := 0; i < numTypes; i++ {
		numStates *= base
	}
	dpFlat := make([]int, numStates*(numTypes+1))
	for i := range dpFlat {
		dpFlat[i] = types.INF
	}

	// Справочники setup/process/due
	setup := make([]int, numTypes+1)
	process := make([]int, numTypes+1)
	due := make([][]int, numTypes+1)
	for t := 1; t <= numTypes; t++ {
		due[t] = make([]int, piecesPerType)
	}
	for _, j := range jobs {
		setup[j.Type] = j.Setup
		process[j.Type] = j.Process
		due[j.Type][j.Instance-1] = j.Deadline
	}

	dpFlat[0*(numTypes+1)+0] = 0
	totalJobs := numTypes * piecesPerType

	// Предвычисляем факторы степеней
	factors := make([]int, numTypes+1)
	exp := 1
	for i := 0; i < numTypes-1; i++ {
		exp *= base
	}
	for t := 1; t <= numTypes; t++ {
		factors[t] = exp
		exp /= base
	}

	// Генерируем коды по слоям
	codesByLayer := make([][]int, totalJobs+1)
	curCounts := make([]int, numTypes)
	var genLayer func(idx, sumSoFar, target int)
	genLayer = func(idx, sumSoFar, target int) {
		if idx == numTypes-1 {
			rem := target - sumSoFar
			if rem >= 0 && rem <= piecesPerType {
				curCounts[idx] = rem
				c := EncodeCountsToCode(curCounts, base)
				codesByLayer[target] = append(codesByLayer[target], c)
			}
			return
		}
		maxV := piecesPerType
		if target-sumSoFar < maxV {
			maxV = target - sumSoFar
		}
		for v := 0; v <= maxV; v++ {
			curCounts[idx] = v
			genLayer(idx+1, sumSoFar+v, target)
		}
	}
	for s := 1; s <= totalJobs; s++ {
		genLayer(0, 0, s)
	}

	// Параллельный DP по слоям
	workers := runtime.NumCPU()
	var wg sync.WaitGroup
	for s := 1; s <= totalJobs; s++ {
		codes := codesByLayer[s]
		if len(codes) == 0 {
			continue
		}
		chunk := (len(codes) + workers - 1) / workers
		for w := 0; w < workers; w++ {
			start := w * chunk
			if start >= len(codes) {
				break
			}
			end := start + chunk
			if end > len(codes) {
				end = len(codes)
			}
			wg.Add(1)
			go func(block []int) {
				defer wg.Done()
				for _, code := range block {
					baseIdx := code * (numTypes + 1)
					for last := 1; last <= numTypes; last++ {
						cnt := (code / factors[last]) % base
						if cnt == 0 {
							continue
						}
						prev := code - factors[last]
						prevIdx := prev * (numTypes + 1)
						best := types.INF
						for pl := 0; pl <= numTypes; pl++ {
							pt := dpFlat[prevIdx+pl]
							if pt == types.INF {
								continue
							}
							var cost int
							if pl == last {
								cost = process[last]
							} else {
								cost = setup[last] + process[last]
							}
							if pt+cost <= due[last][cnt-1] && pt+cost < best {
								best = pt + cost
							}
						}
						if best < types.INF {
							dpFlat[baseIdx+last] = best
						}
					}
				}
			}(codes[start:end])
		}
		wg.Wait()
	}

	return dpFlat, base
}

// EncodeCountsToCode кодирует вектор counts в целое при основании base.
func EncodeCountsToCode(counts []int, base int) int {
	code := 0
	for _, v := range counts {
		code = code*base + v
	}
	return code
}

// ReconstructSchedule делает backtracking по dp и возвращает упорядоченный список работ.
func ReconstructSchedule(dpFlat []int, jobs []types.Job, numTypes, piecesPerType, base, bestLast int) []types.Job {
	// Подготовка справочников
	setup := make([]int, numTypes+1)
	process := make([]int, numTypes+1)
	due := make([][]int, numTypes+1)
	for t := 1; t <= numTypes; t++ {
		due[t] = make([]int, piecesPerType)
	}
	for _, j := range jobs {
		setup[j.Type] = j.Setup
		process[j.Type] = j.Process
		due[j.Type][j.Instance-1] = j.Deadline
	}

	// Факторы степеней
	factors := make([]int, numTypes+1)
	exp := 1
	for i := 0; i < numTypes-1; i++ {
		exp *= base
	}
	for t := 1; t <= numTypes; t++ {
		factors[t] = exp
		exp /= base
	}

	total := numTypes * piecesPerType
	counts := make([]int, numTypes+1)
	for i := 1; i <= numTypes; i++ {
		counts[i] = piecesPerType
	}
	code := EncodeCountsToCode(counts[1:], base)
	last := bestLast

	var rev []types.Job
	for step := total; step > 0; step-- {
		inst := counts[last]
		rev = append(
			rev, types.Job{
				Type:     last,
				Instance: inst,
				Setup:    setup[last],
				Process:  process[last],
				Deadline: due[last][inst-1],
			},
		)
		counts[last]--
		prevCode := code - factors[last]
		curT := dpFlat[code*(numTypes+1)+last]
		found := -1
		for pl := 0; pl <= numTypes; pl++ {
			pt := dpFlat[prevCode*(numTypes+1)+pl]
			if pt == types.INF {
				continue
			}
			var cost int
			if pl == last {
				cost = process[last]
			} else {
				cost = setup[last] + process[last]
			}
			if pt+cost == curT {
				found = pl
				break
			}
		}
		last = found
		code = prevCode
	}
	// Переворачиваем
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

// BuildStepsFromSchedule преобразует расписание работ в шаги для визуализации.
func BuildStepsFromSchedule(schedule []types.Job) ([]types.Step, []int) {
	var steps []types.Step
	var ticks []int
	var timeCursor int
	var prevType = -1

	for _, job := range schedule {
		setup := 0
		if job.Type != prevType {
			setup = job.Setup
		}

		step := types.Step{
			Type:     job.Type,
			Instance: job.Instance,
			Start:    timeCursor,
			SetupLen: setup,
			ProcLen:  job.Process,
		}
		steps = append(steps, step)

		end := timeCursor + setup + job.Process
		ticks = append(ticks, end)

		timeCursor = end
		prevType = job.Type
	}

	return steps, ticks
}
