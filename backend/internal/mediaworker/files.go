package mediaworker

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func saveUploadedFile(dir string, fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	path := filepath.Join(dir, filepath.Base(fh.Filename))
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return path, err
}

func sendFile(w http.ResponseWriter, path, contentType string) {
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}
