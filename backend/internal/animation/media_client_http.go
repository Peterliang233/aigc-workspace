package animation

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func (c *MediaClient) uploadOne(ctx context.Context, path, field, inputPath, outputPath string, fields map[string]string) error {
	return c.uploadMany(ctx, path, field, []string{inputPath}, outputPath, fields)
}

func (c *MediaClient) uploadMany(ctx context.Context, path, field string, inputs []string, outputPath string, fields map[string]string) error {
	if !c.Enabled() {
		return fmt.Errorf("media worker is not configured")
	}
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, pr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	go func() {
		err := writeMultipartBody(mw, field, inputs, fields)
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

func writeMultipartBody(mw *multipart.Writer, field string, inputs []string, fields map[string]string) error {
	defer mw.Close()
	for key, value := range fields {
		if err := mw.WriteField(key, value); err != nil {
			return err
		}
	}
	for _, path := range inputs {
		if err := writeMultipartFile(mw, field, path); err != nil {
			return err
		}
	}
	return nil
}

func writeMultipartFile(mw *multipart.Writer, field, path string) error {
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
