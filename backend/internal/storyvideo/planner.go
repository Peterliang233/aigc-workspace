package storyvideo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
)

type Planner struct {
	provider   string
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewPlanner(provider, baseURL, apiKey, model string) *Planner {
	provider = strings.ToLower(strings.TrimSpace(provider))
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" && provider == "bltcy" {
		baseURL = "https://api.bltcy.ai"
	}
	if strings.TrimSpace(model) == "" {
		model = "gpt-5.4"
	}
	return &Planner{
		provider:   provider,
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

func (p *Planner) Enabled() bool        { return p != nil && p.baseURL != "" && p.apiKey != "" }
func (p *Planner) DefaultModel() string { return p.model }

func (p *Planner) Draft(ctx context.Context, model string, req DraftRequest) (Draft, error) {
	if strings.TrimSpace(model) == "" {
		model = p.model
	}
	body, _ := plannerRequestBody(model, req)
	url := p.baseURL + "/v1/chat/completions"
	name := p.provider + "_planner"
	logging.DownstreamRequestRaw("story_video_planner_request", name, http.MethodPost, url, map[string]any{
		"model": model, "keywords": req.Keywords, "duration_seconds": req.DurationSeconds,
	}, "application/json", body)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponseRaw("story_video_planner_response", name, http.MethodPost, url, 0, time.Since(start), err, "", nil)
		return Draft{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponseRaw("story_video_planner_response", name, http.MethodPost, url, resp.StatusCode, time.Since(start), fmt.Errorf("bad status"), resp.Header.Get("Content-Type"), raw)
		return Draft{}, fmt.Errorf("planner API error: status=%d body=%s", resp.StatusCode, string(raw))
	}
	logging.DownstreamResponseRaw("story_video_planner_response", name, http.MethodPost, url, resp.StatusCode, time.Since(start), nil, resp.Header.Get("Content-Type"), raw)
	out, err := parseDraftResponse(raw)
	if err != nil {
		return Draft{}, err
	}
	if !hasDraftSignal(out) {
		return Draft{}, errors.New("invalid planner draft: no usable content")
	}
	out = repairDraft(out, req)
	out = normalizeDraft(out, req.DurationSeconds)
	if err := validateDraft(out); err != nil {
		return Draft{}, err
	}
	return out, nil
}

func plannerRequestBody(model string, req DraftRequest) ([]byte, error) {
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": plannerSystemPrompt()},
			{"role": "user", "content": plannerUserPrompt(req)},
		},
	}
	if strings.EqualFold(strings.TrimSpace(model), "gpt-5.4") || strings.Contains(strings.ToLower(model), "gpt") {
		payload["response_format"] = map[string]any{"type": "json_object"}
	}
	return json.Marshal(payload)
}

func normalizeDraft(in Draft, totalSeconds int) Draft {
	if len(in.Shots) == 0 {
		return in
	}
	target := totalSeconds * 1000
	if target <= 0 {
		target = len(in.Shots) * 4000
	}
	sum := 0
	for i := range in.Shots {
		if in.Shots[i].DurationMS <= 0 {
			in.Shots[i].DurationMS = 4000
		}
		sum += in.Shots[i].DurationMS
	}
	if sum <= 0 || sum == target {
		return in
	}
	for i := range in.Shots {
		in.Shots[i].DurationMS = maxInt(1000, in.Shots[i].DurationMS*target/sum)
	}
	return in
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func validateDraft(in Draft) error {
	if strings.TrimSpace(in.Title) == "" {
		return errors.New("invalid planner draft: empty title")
	}
	if strings.TrimSpace(in.Summary) == "" {
		return errors.New("invalid planner draft: empty summary")
	}
	if strings.TrimSpace(in.ScriptText) == "" {
		return errors.New("invalid planner draft: empty script_text")
	}
	if strings.TrimSpace(in.NarrationText) == "" {
		return errors.New("invalid planner draft: empty narration_text")
	}
	if len(in.Shots) == 0 {
		return errors.New("invalid planner draft: empty shots")
	}
	for i, shot := range in.Shots {
		if strings.TrimSpace(shot.Title) == "" {
			return fmt.Errorf("invalid planner draft: shot %d empty title", i+1)
		}
		if strings.TrimSpace(shot.StoryBeat) == "" {
			return fmt.Errorf("invalid planner draft: shot %d empty story_beat", i+1)
		}
		if strings.TrimSpace(shot.NarrationLine) == "" {
			return fmt.Errorf("invalid planner draft: shot %d empty narration_line", i+1)
		}
		if strings.TrimSpace(shot.ImagePrompt) == "" {
			return fmt.Errorf("invalid planner draft: shot %d empty image_prompt", i+1)
		}
		if shot.DurationMS <= 0 {
			return fmt.Errorf("invalid planner draft: shot %d invalid duration_ms", i+1)
		}
	}
	return nil
}
