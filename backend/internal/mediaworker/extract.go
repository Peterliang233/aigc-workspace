package mediaworker

import (
	"net/http"
	"os"
	"path/filepath"
)

func extractLastFrame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmpDir, err := os.MkdirTemp("", "media-frame-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)
	if err := r.ParseMultipartForm(350 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	inputPath, err := saveUploadedFile(tmpDir, files[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	outputPath := filepath.Join(tmpDir, "last.jpg")
	err = runFFmpeg("-y", "-sseof", "-0.1", "-i", inputPath, "-frames:v", "1", "-q:v", "2", outputPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	sendFile(w, outputPath, "image/jpeg")
}
