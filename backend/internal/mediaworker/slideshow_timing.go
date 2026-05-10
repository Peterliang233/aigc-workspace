package mediaworker

import "strings"

const narrationTailPaddingMS = 1500

func alignSlideshowDurations(durations []int, audioPath string) ([]int, int, error) {
	if strings.TrimSpace(audioPath) == "" {
		return append([]int(nil), durations...), 0, nil
	}
	audioMS, err := probeMediaDurationMS(audioPath)
	if err != nil {
		return nil, 0, err
	}
	return stretchDurations(durations, audioMS+narrationTailPaddingMS), audioMS, nil
}

func stretchDurations(durations []int, targetMS int) []int {
	total := 0
	for _, d := range durations {
		total += d
	}
	if len(durations) == 0 || total == targetMS {
		return append([]int(nil), durations...)
	}
	adjusted := make([]int, len(durations))
	remaining := targetMS
	for i, d := range durations {
		if i == len(durations)-1 {
			adjusted[i] = maxInt(1000, remaining)
			break
		}
		scaled := int(int64(d) * int64(targetMS) / int64(total))
		adjusted[i] = maxInt(1000, scaled)
		remaining -= adjusted[i]
	}
	return adjusted
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
