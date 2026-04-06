package animation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aigc-backend/internal/logging"
)

type PromptPlanner struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewPromptPlanner(baseURL, apiKey, model string) *PromptPlanner {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.bltcy.ai"
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gpt-5.4"
	}
	return &PromptPlanner{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

func (p *PromptPlanner) Enabled() bool { return p != nil && p.apiKey != "" && p.baseURL != "" }
func (p *PromptPlanner) Model() string { return p.model }

func (p *PromptPlanner) Plan(ctx context.Context, model string, req PromptPlanRequest) ([]PlannedSegment, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		model = p.model
	}
	body, _ := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": plannerSystemPrompt()},
			{"role": "user", "content": plannerUserPrompt(req)},
		},
		"response_format": map[string]any{
			"type":        "json_schema",
			"json_schema": plannerSchema(),
		},
	})
	url := p.baseURL + "/v1/chat/completions"
	logging.DownstreamRequest("planner_request", "bltcy_planner", http.MethodPost, url, map[string]any{
		"model":            model,
		"segment_count":    len(req.SegmentDurations),
		"segment_lengths":  req.SegmentDurations,
		"total_duration_s": req.TotalSeconds,
		"prompt":           logging.DownstreamPrompt(req.Prompt),
	})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	hreq.Header.Set("Authorization", "Bearer "+p.apiKey)
	hreq.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := p.httpClient.Do(hreq)
	if err != nil {
		logging.DownstreamResponse("planner_response", "bltcy_planner", http.MethodPost, url, 0, time.Since(start), err)
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.DownstreamResponse("planner_response", "bltcy_planner", http.MethodPost, url, resp.StatusCode, time.Since(start), fmt.Errorf("bad status"))
		return nil, fmt.Errorf("planner API error: status=%d body=%s", resp.StatusCode, string(raw))
	}
	logging.DownstreamResponse("planner_response", "bltcy_planner", http.MethodPost, url, resp.StatusCode, time.Since(start), nil)
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if len(payload.Choices) == 0 || strings.TrimSpace(payload.Choices[0].Message.Content) == "" {
		return nil, fmt.Errorf("planner returned empty content")
	}
	var out plannerResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(payload.Choices[0].Message.Content)), &out); err != nil {
		return nil, err
	}
	if len(out.Segments) != len(req.SegmentDurations) {
		return nil, fmt.Errorf("planner returned %d segments, expected %d", len(out.Segments), len(req.SegmentDurations))
	}
	segments := make([]PlannedSegment, 0, len(out.Segments))
	for idx, seg := range out.Segments {
		prompt := strings.TrimSpace(seg.Prompt)
		if prompt == "" {
			return nil, fmt.Errorf("planner returned empty prompt for segment %d", idx+1)
		}
		segments = append(segments, PlannedSegment{
			DurationSeconds: req.SegmentDurations[idx],
			Prompt:          prompt,
			Continuity:      strings.TrimSpace(seg.Continuity),
		})
	}
	return segments, nil
}
