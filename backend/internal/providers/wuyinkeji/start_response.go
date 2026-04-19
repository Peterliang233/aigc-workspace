package wuyinkeji

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseStartJobID(raw []byte) (string, error) {
	var out startResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	jobID := strings.TrimSpace(out.Data.ID)
	if jobID != "" {
		return jobID, nil
	}
	msg := strings.TrimSpace(out.Msg)
	if msg != "" {
		if out.Code != 0 {
			return "", fmt.Errorf("wuyinkeji start failed: code=%d msg=%s", out.Code, msg)
		}
		return "", fmt.Errorf("wuyinkeji start failed: %s", msg)
	}
	return "", fmt.Errorf("wuyinkeji start missing id: %s", string(raw))
}
