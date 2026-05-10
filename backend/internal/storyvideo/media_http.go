package storyvideo

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (c *MediaClient) uploadCompose(ctx context.Context, path string, images []string, audioPath, outputPath string, fields map[string]string) error {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, pr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	go func() {
		err := writeComposeMultipart(mw, images, audioPath, fields)
		_ = pw.CloseWithError(err)
	}()
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("media worker error: status=%d body=%s", resp.StatusCode, string(b))
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func writeComposeMultipart(mw *multipart.Writer, images []string, audioPath string, fields map[string]string) error {
	defer mw.Close()
	for key, value := range fields {
		if err := mw.WriteField(key, value); err != nil {
			return err
		}
	}
	for _, image := range images {
		if err := writeComposeFile(mw, "images", image); err != nil {
			return err
		}
	}
	if strings.TrimSpace(audioPath) != "" {
		if err := writeComposeFile(mw, "audio", audioPath); err != nil {
			return err
		}
	}
	return nil
}

func writeComposeFile(mw *multipart.Writer, field, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	part, err := mw.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	return err
}

func intsJSON(values []int) string {
	if len(values) == 0 {
		return "[]"
	}
	var out strings.Builder
	out.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			out.WriteByte(',')
		}
		out.WriteString(strconv.Itoa(value))
	}
	out.WriteByte(']')
	return out.String()
}
