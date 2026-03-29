package gptbest

import (
	"encoding/json"
	"fmt"
	"strings"
)

func normalizeVideoStatus(in string) string {
	s := strings.TrimSpace(in)
	if s == "" {
		return ""
	}
	u := strings.ToUpper(s)
	switch u {
	case "SUCCEEDED", "SUCCESS", "COMPLETED", "DONE", "FINISHED":
		return "succeeded"
	case "FAILED", "FAIL", "ERROR", "REJECTED", "TIMEOUT", "EXPIRED", "ABORTED", "CANCELED", "CANCELLED":
		return "failed"
	case "PENDING", "QUEUED", "SUBMITTED", "CREATED":
		return "queued"
	case "RUNNING", "IN_PROGRESS", "INPROGRESS", "PROCESSING", "GENERATING":
		return "running"
	}
	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "succeed"), strings.Contains(low, "complete"), strings.Contains(low, "finish"):
		return "succeeded"
	case strings.Contains(low, "fail"), strings.Contains(low, "error"), strings.Contains(low, "reject"), strings.Contains(low, "cancel"):
		return "failed"
	case strings.Contains(low, "queue"), strings.Contains(low, "pending"), strings.Contains(low, "submit"):
		return "queued"
	default:
		return "running"
	}
}

func pickJSONStr(raw []byte, keys ...string) string {
	var m any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	return findStringByKeys(m, keys...)
}

func findStringByKeys(v any, keys ...string) string {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range keys {
			for mk, mv := range t {
				if strings.EqualFold(strings.TrimSpace(mk), strings.TrimSpace(k)) {
					if s := scalarToString(mv); s != "" {
						return s
					}
				}
			}
		}
		for _, mv := range t {
			if s := findStringByKeys(mv, keys...); s != "" {
				return s
			}
		}
	case []any:
		for _, it := range t {
			if s := findStringByKeys(it, keys...); s != "" {
				return s
			}
		}
	}
	return ""
}

func scalarToString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case bool, float64, float32, int, int64, int32, uint, uint64, uint32:
		return strings.TrimSpace(fmt.Sprint(x))
	default:
		return ""
	}
}

func pickVideoURL(raw []byte) string {
	var m any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	for _, k := range []string{"video_url", "output_url", "result_url", "play_url", "url"} {
		if s := findStringByKeys(m, k); isVideoURL(s) {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func parseVideoStatusAndError(raw []byte) (string, string) {
	statusRaw := pickJSONStr(raw, "status", "state", "task_status", "phase")
	jobErr := pickJSONStr(raw, "error", "error_message", "reason", "failed_reason", "fail_reason", "message")
	status := normalizeVideoStatus(statusRaw)
	if status == "" {
		if looksFailedMessage(jobErr) {
			status = "failed"
		} else {
			status = "running"
		}
	}
	if status == "failed" && jobErr == "" {
		jobErr = "任务失败"
	}
	return status, strings.TrimSpace(jobErr)
}

func looksFailedMessage(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return false
	}
	for _, kw := range []string{"fail", "error", "denied", "invalid", "timeout", "cancel", "not found", "失败", "错误"} {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func isVideoURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	u := strings.ToLower(s)
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "/") {
		return strings.Contains(u, ".mp4") || strings.Contains(u, ".mov") || strings.Contains(u, "video") || strings.Contains(u, "/api/assets/")
	}
	return false
}
