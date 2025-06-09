package gms

import (
	"fmt"
	"sort"
	"strings"

	"dp-calc/types"
)

// BuildGMS строит GAMS-модель для задачи планирования с заданными параметрами.
func BuildGMS(jobs []types.Job, T, P int) []byte {
	var buf strings.Builder
	buf.WriteString("Sets\n")
	buf.WriteString(fmt.Sprintf("   i /i1*i%d/\n", T))
	buf.WriteString(fmt.Sprintf("   j 'Jobs' /J1*J%d/\n", P))
	buf.WriteString(fmt.Sprintf("   k 'Setup/Process' /k1*k%d/\n", T*P))
	buf.WriteString("Alias(i,l);\n")
	buf.WriteString("Alias(k,t);\n")
	if P == 3 {
		buf.WriteString("Alias(j,j1);\n")
	}
	buf.WriteString("\nParameters\n")
	sort.SliceStable(
		jobs, func(a, b int) bool {
			if jobs[a].Type == jobs[b].Type {
				return jobs[a].Instance < jobs[b].Instance
			}
			return jobs[a].Type < jobs[b].Type
		},
	)
	buf.WriteString("   p(i) /\n")
	for i := 0; i < T; i++ {
		buf.WriteString(fmt.Sprintf("     i%d=%d", i+1, jobs[i*P].Process))
		if i < T-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString(" /\n")
		}
	}
	buf.WriteString("\n   s(i) /\n")
	for i := 0; i < T; i++ {
		buf.WriteString(fmt.Sprintf("     i%d=%d", i+1, jobs[i*P].Setup))
		if i < T-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString(" /\n")
		}
	}
	buf.WriteString("\n   d(i,j) /\n")
	for i := 0; i < T; i++ {
		for j := 0; j < P; j++ {
			idx := i*P + j
			buf.WriteString(fmt.Sprintf("     i%d.J%d = %d", i+1, j+1, jobs[idx].Deadline))
			if j < P-1 {
				buf.WriteString(",\n")
			}
		}
		if i < T-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString(" /\n")
		}
	}
	buf.WriteString(";\n")
	return []byte(buf.String())
}
