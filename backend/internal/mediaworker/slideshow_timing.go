package mediaworker

import "strings"

const narrationTailPaddingMS = 600

func alignSlideshowDurations(durations []int, audioPath string) ([]int, error) {
	if strings.TrimSpace(audioPath) == "" {
		return append([]int(nil), durations...), nil
	}
	audioMS, err := probeMediaDurationMS(audioPath)
	if err != nil {
		return nil, err
	}
	return stretchDurations(durations, audioMS+narrationTailPaddingMS), nil
}

func stretchDurations(durations []int, targetMS int) []int {
	total := 0
	for _, duration := range durations {
		total += duration
	}
	if len(durations) == 0 || total >= targetMS {
		return append([]int(nil), durations...)
	}
	extra := targetMS - total
	adjusted := make([]int, len(durations))
	remainingBase, remainingExtra := total, extra
	for i, duration := range durations {
		add := remainingExtra
		if i < len(durations)-1 && remainingBase > 0 {
			add = int((int64(remainingExtra)*int64(duration) + int64(remainingBase)/2) / int64(remainingBase))
		}
		adjusted[i] = duration + add
		remainingBase -= duration
		remainingExtra -= add
	}
	return adjusted
}
