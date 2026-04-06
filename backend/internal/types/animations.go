package types

type AnimationJobCreateRequest struct {
	Provider        string `json:"provider,omitempty"`
	Model           string `json:"model,omitempty"`
	PlannerModel    string `json:"planner_model,omitempty"`
	Prompt          string `json:"prompt"`
	DurationSeconds int    `json:"duration_seconds"`
	AspectRatio     string `json:"aspect_ratio,omitempty"`
	LeadImage       string `json:"lead_image,omitempty"`
	Seed            *int64 `json:"seed,omitempty"`
}

type AnimationJobCreateResponse struct {
	JobID           string `json:"job_id"`
	Status          string `json:"status"`
	Provider        string `json:"provider"`
	Model           string `json:"model,omitempty"`
	DurationSeconds int    `json:"duration_seconds"`
}

type AnimationSegmentResponse struct {
	Index          int    `json:"index"`
	Status         string `json:"status"`
	Duration       int    `json:"duration_seconds"`
	Prompt         string `json:"prompt,omitempty"`
	Continuity     string `json:"continuity,omitempty"`
	SourceJobID    string `json:"source_job_id,omitempty"`
	VideoURL       string `json:"video_url,omitempty"`
	LastFrameReady bool   `json:"last_frame_ready,omitempty"`
	Error          string `json:"error,omitempty"`
}

type AnimationJobGetResponse struct {
	JobID             string                     `json:"job_id"`
	Status            string                     `json:"status"`
	Provider          string                     `json:"provider"`
	Model             string                     `json:"model,omitempty"`
	Prompt            string                     `json:"prompt,omitempty"`
	DurationSeconds   int                        `json:"duration_seconds"`
	PlannerStatus     string                     `json:"planner_status,omitempty"`
	PlannerModel      string                     `json:"planner_model,omitempty"`
	PlannerError      string                     `json:"planner_error,omitempty"`
	SegmentCount      int                        `json:"segment_count"`
	CompletedSegments int                        `json:"completed_segments"`
	CurrentSegment    int                        `json:"current_segment"`
	VideoURL          string                     `json:"video_url,omitempty"`
	Error             string                     `json:"error,omitempty"`
	Segments          []AnimationSegmentResponse `json:"segments,omitempty"`
}
