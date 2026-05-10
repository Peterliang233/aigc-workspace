package mediaworker

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func concatAudios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmpDir, err := os.MkdirTemp("", "media-audio-concat-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)
	if err := r.ParseMultipartForm(700 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["audios"]
	if len(files) == 0 {
		http.Error(w, "missing audios", http.StatusBadRequest)
		return
	}
	inputs, durations := make([]string, 0, len(files)), make([]int, 0, len(files))
	for _, file := range files {
		path, err := saveUploadedFile(tmpDir, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		durationMS, err := probeMediaDurationMS(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		inputs = append(inputs, path)
		durations = append(durations, durationMS)
	}
	listPath := filepath.Join(tmpDir, "concat.txt")
	if err := os.WriteFile(listPath, []byte(buildConcatList(inputs)), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	outputPath := filepath.Join(tmpDir, "audio.m4a")
	if err := runFFmpeg("-y", "-f", "concat", "-safe", "0", "-i", listPath, "-vn", "-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart", outputPath); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	totalMS, err := probeMediaDurationMS(outputPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	rawDurations, _ := json.Marshal(durations)
	w.Header().Set("X-Audio-Durations-Ms", string(rawDurations))
	w.Header().Set("X-Audio-Duration-Ms", strconv.Itoa(totalMS))
	sendFile(w, outputPath, "audio/mp4")
}
