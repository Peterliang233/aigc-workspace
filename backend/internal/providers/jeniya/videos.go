package jeniya

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"aigc-backend/internal/types"
)

func (p *Provider) StartVideoJob(ctx context.Context, req types.VideoJobCreateRequest) (string, error) {
	if p.baseURL == "" || p.apiKey == "" {
		return "", errors.New("平台未配置 Base URL 或 API Key")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		return "", errors.New("model is required")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return "", errors.New("prompt is required")
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("model", model)
	_ = writer.WriteField("prompt", prompt)
	_ = writer.WriteField("seconds", strconv.Itoa(pickDuration(req.DurationSeconds)))
	_ = writer.WriteField("size", pickSize(req))
	if image := strings.TrimSpace(req.Image); image != "" {
		name, buf, err := p.loadImage(ctx, image)
		if err != nil {
			return "", err
		}
		part, err := writer.CreateFormFile("input_reference", name)
		if err != nil {
			return "", err
		}
		if _, err = part.Write(buf); err != nil {
			return "", err
		}
	}
	for key, value := range req.Extra {
		if skipVideoField(key) {
			continue
		}
		_ = writer.WriteField(strings.TrimSpace(key), fmt.Sprint(value))
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	url := p.baseURL + "/v1/video/create"
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := p.http.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("jeniya video create failed: status=%d body=%s", resp.StatusCode, string(raw))
	}
	jobID := findString(raw, "id", "video_id", "task_id")
	if jobID == "" {
		return "", errors.New("video create returned empty id")
	}
	return jobID, nil
}

func (p *Provider) GetVideoJob(ctx context.Context, jobID string) (string, string, string, error) {
	if p.baseURL == "" || p.apiKey == "" {
		return "", "", "", errors.New("平台未配置 Base URL 或 API Key")
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return "", "", "", errors.New("job_id is required")
	}
	raw, err := p.getJSON(ctx, "/v1/videos/"+jobID)
	if err != nil {
		return "", "", "", err
	}
	status := normalizeStatus(findString(raw, "status", "state"))
	jobErr := findString(raw, "error", "message", "failed_reason", "failure_reason")
	videoURL := findString(raw, "video_url", "url")
	if status == "succeeded" && strings.TrimSpace(videoURL) == "" {
		content, err := p.getJSON(ctx, "/v1/videos/"+jobID+"/content")
		if err == nil {
			videoURL = findString(content, "video_url", "download_url", "url")
		}
	}
	return status, strings.TrimSpace(videoURL), strings.TrimSpace(jobErr), nil
}

func (p *Provider) getJSON(ctx context.Context, path string) ([]byte, error) {
	url := p.baseURL + path
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jeniya %s failed: status=%d body=%s", path, resp.StatusCode, string(raw))
	}
	return raw, nil
}

func pickDuration(v int) int {
	if v > 0 {
		return v
	}
	return 5
}

func pickSize(req types.VideoJobCreateRequest) string {
	if s := strings.TrimSpace(req.ImageSize); s != "" {
		return s
	}
	if req.Extra != nil {
		if s := strings.TrimSpace(fmt.Sprint(req.Extra["size"])); s != "" && s != "<nil>" {
			return s
		}
	}
	switch strings.TrimSpace(req.AspectRatio) {
	case "9:16":
		return "720x1280"
	case "1:1":
		return "1024x1024"
	default:
		return "1280x720"
	}
}

func skipVideoField(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "", "model", "prompt", "seconds", "size", "input_reference":
		return true
	default:
		return false
	}
}
