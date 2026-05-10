package mediaworker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func composeSlideshow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmpDir, err := os.MkdirTemp("", "media-slideshow-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)
	if err := r.ParseMultipartForm(700 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	images, err := slideshowInputs(tmpDir, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	durations, err := slideshowDurations(r.FormValue("durations_ms_json"), len(images))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	width, height := slideshowSize(r.FormValue("aspect_ratio"))
	audio := ""
	if files := r.MultipartForm.File["audio"]; len(files) > 0 {
		audio, err = saveUploadedFile(tmpDir, files[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	var audioMS int
	durations, audioMS, err = alignSlideshowDurations(durations, audio)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	clips := make([]string, 0, len(images))
	for i, image := range images {
		clipPath := filepath.Join(tmpDir, fmt.Sprintf("clip-%02d.mp4", i+1))
		if err := renderSlideshowClip(image, clipPath, width, height, durations[i]); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		clips = append(clips, clipPath)
	}
	merged := filepath.Join(tmpDir, "merged.mp4")
	if err := os.WriteFile(filepath.Join(tmpDir, "concat.txt"), []byte(buildConcatList(clips)), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := runFFmpeg("-y", "-f", "concat", "-safe", "0", "-i", filepath.Join(tmpDir, "concat.txt"), "-c:v", "copy", merged); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	output := filepath.Join(tmpDir, "output.mp4")
	if err := muxSlideshow(merged, audio, output, audioMS); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	sendFile(w, output, "video/mp4")
}

func slideshowInputs(dir string, r *http.Request) ([]string, error) {
	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		return nil, fmt.Errorf("missing images")
	}
	out := make([]string, 0, len(files))
	for _, fh := range files {
		path, err := saveUploadedFile(dir, fh)
		if err != nil {
			return nil, err
		}
		out = append(out, path)
	}
	return out, nil
}

func slideshowDurations(raw string, count int) ([]int, error) {
	var values []int
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &values); err != nil {
		return nil, fmt.Errorf("invalid durations_ms_json")
	}
	if len(values) != count {
		return nil, fmt.Errorf("durations count mismatch")
	}
	for i := range values {
		if values[i] < 1000 {
			values[i] = 1000
		}
	}
	return values, nil
}

func slideshowSize(aspect string) (int, int) {
	switch strings.TrimSpace(aspect) {
	case "9:16":
		return 720, 1280
	case "1:1":
		return 1024, 1024
	case "4:3":
		return 1024, 768
	case "3:4":
		return 768, 1024
	default:
		return 1280, 720
	}
}

func renderSlideshowClip(imagePath, outputPath string, width, height, durationMS int) error {
	seconds := strconv.FormatFloat(float64(durationMS)/1000, 'f', 2, 64)
	filter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,format=yuv420p", width, height, width, height)
	return runFFmpeg("-y", "-loop", "1", "-i", imagePath, "-t", seconds, "-vf", filter, "-r", "25", "-c:v", "libx264", "-preset", "veryfast", "-pix_fmt", "yuv420p", outputPath)
}

func muxSlideshow(videoPath, audioPath, outputPath string, audioDurationMS int) error {
	if strings.TrimSpace(audioPath) == "" {
		return runFFmpeg("-y", "-i", videoPath, "-c:v", "copy", outputPath)
	}
	duration := strconv.FormatFloat(float64(audioDurationMS)/1000+1.0, 'f', 2, 64)
	return runFFmpeg("-y",
		"-stream_loop", "-1", "-i", videoPath,
		"-i", audioPath,
		"-t", duration,
		"-c:v", "copy", "-c:a", "aac",
		"-movflags", "+faststart", outputPath)
}
