package mediaworker

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func concatVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmpDir, err := os.MkdirTemp("", "media-concat-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)
	if err := r.ParseMultipartForm(700 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var inputs []string
	for _, fh := range r.MultipartForm.File["files"] {
		path, err := saveUploadedFile(tmpDir, fh)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		inputs = append(inputs, path)
	}
	if len(inputs) == 0 {
		http.Error(w, "missing files", http.StatusBadRequest)
		return
	}
	listPath := filepath.Join(tmpDir, "concat.txt")
	if err := os.WriteFile(listPath, []byte(buildConcatList(inputs)), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outputPath := filepath.Join(tmpDir, "output.mp4")
	args := []string{"-y", "-f", "concat", "-safe", "0", "-i", listPath, "-map", "0:v:0", "-map", "0:a?", "-c:v", "libx264", "-preset", "veryfast", "-pix_fmt", "yuv420p", "-c:a", "aac", "-movflags", "+faststart"}
	if raw := strings.TrimSpace(r.FormValue("duration_seconds")); raw != "" {
		if _, err := strconv.Atoi(raw); err == nil {
			args = append(args, "-t", raw)
		}
	}
	args = append(args, outputPath)
	if err := runFFmpeg(args...); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	sendFile(w, outputPath, "video/mp4")
}

func buildConcatList(inputs []string) string {
	var lines []string
	for _, path := range inputs {
		lines = append(lines, "file '"+strings.ReplaceAll(path, "'", "'\\''")+"'")
	}
	return strings.Join(lines, "\n")
}
