package mediaworker

import "strings"

const narrationEndHoldMS = 3000

func alignSlideshowDurations(durations []int, audioPath string) ([]int, int, int, error) {
	if strings.TrimSpace(audioPath) == "" {
		out := append([]int(nil), durations...)
		return out, 0, durationSum(out), nil
	}
	audioMS, err := probeMediaDurationMS(audioPath)
	if err != nil {
		return nil, 0, 0, err
	}
	out := alignDurationsToNarrationAudio(durations, audioMS)
	return out, audioMS, durationSum(out), nil
}

func alignDurationsToNarrationAudio(durations []int, audioMS int) []int {
	out := stretchDurations(durations, audioMS)
	if len(out) == 0 {
		return out
	}
	out[len(out)-1] += narrationEndHoldMS
	return out
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

func durationSum(durations []int) int {
	total := 0
	for _, d := range durations {
		total += d
	}
	return total
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
