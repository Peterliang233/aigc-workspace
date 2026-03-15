package siliconflow

type sfVideoSubmitRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	ImageSize      string `json:"image_size,omitempty"` // 1280x720|720x1280|960x960
	Image          string `json:"image,omitempty"`      // URL or base64
	Seed           *int64 `json:"seed,omitempty"`
}

type sfVideoSubmitResponse struct {
	RequestID string `json:"requestId"`
}

type sfVideoStatusRequest struct {
	RequestID string `json:"requestId"`
}

type sfVideoStatusResponse struct {
	Status  string `json:"status"` // InQueue|InProgress|Succeed|Failed
	Reason  string `json:"reason"`
	Results struct {
		Videos []struct {
			URL string `json:"url"`
		} `json:"videos"`
	} `json:"results"`
}
