package animation

import (
	"errors"
	"sort"
)

func BuildPlan(total int, allowed []int) ([]int, int, error) {
	allowed = normalizeDurations(allowed)
	if total <= 0 {
		return nil, 0, errors.New("duration_seconds must be greater than 0")
	}
	if len(allowed) == 0 {
		return nil, 0, errors.New("no valid segment durations")
	}

	limit := total + allowed[len(allowed)-1]
	best := make([]int, limit+1)
	prev := make([]int, limit+1)
	for i := 1; i <= limit; i++ {
		best[i] = 1 << 30
		prev[i] = -1
	}

	for sum := 0; sum <= limit; sum++ {
		if sum > 0 && best[sum] == 1<<30 {
			continue
		}
		for _, d := range allowed {
			next := sum + d
			if next > limit {
				continue
			}
			if best[sum]+1 < best[next] {
				best[next] = best[sum] + 1
				prev[next] = d
			}
		}
	}

	target := -1
	for sum := total; sum <= limit; sum++ {
		if prev[sum] > 0 {
			target = sum
			break
		}
	}
	if target < 0 {
		return nil, 0, errors.New("unable to build segment plan")
	}

	var plan []int
	for target > 0 {
		d := prev[target]
		if d <= 0 {
			return nil, 0, errors.New("invalid segment reconstruction")
		}
		plan = append(plan, d)
		target -= d
	}
	sort.Sort(sort.Reverse(sort.IntSlice(plan)))
	totalPlanned := 0
	for _, d := range plan {
		totalPlanned += d
	}
	return plan, totalPlanned, nil
}

func normalizeDurations(in []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(in))
	for _, v := range in {
		if v <= 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}
