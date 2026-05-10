package mediaworker

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

func probeMediaDurationMS(path string) (int, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("invalid ffprobe duration: %q", strings.TrimSpace(string(out)))
	}
	return int(math.Ceil(seconds * 1000)), nil
}
