package types

type StoryVideoDraftRequest struct {
	Keywords        []string `json:"keywords"`
	Theme           string   `json:"theme,omitempty"`
	Audience        string   `json:"audience,omitempty"`
	Tone            string   `json:"tone,omitempty"`
	Extra           string   `json:"extra,omitempty"`
	DurationSeconds int      `json:"duration_seconds"`
	AspectRatio     string   `json:"aspect_ratio,omitempty"`
	PlannerProvider string   `json:"planner_provider,omitempty"`
	PlannerModel    string   `json:"planner_model,omitempty"`
	ImageProvider   string   `json:"image_provider,omitempty"`
	ImageModel      string   `json:"image_model,omitempty"`
	AudioProvider   string   `json:"audio_provider,omitempty"`
	AudioModel      string   `json:"audio_model,omitempty"`
	AudioVoice      string   `json:"audio_voice,omitempty"`
}

type StoryVideoDraftShotInput struct {
	ID            string `json:"id,omitempty"`
	Title         string `json:"title"`
	StoryBeat     string `json:"story_beat"`
	NarrationLine string `json:"narration_line"`
	ImagePrompt   string `json:"image_prompt"`
	DurationMS    int    `json:"duration_ms"`
}

type StoryVideoDraftUpdateRequest struct {
	Keywords        []string                   `json:"keywords"`
	Theme           string                     `json:"theme,omitempty"`
	Audience        string                     `json:"audience,omitempty"`
	Tone            string                     `json:"tone,omitempty"`
	Extra           string                     `json:"extra,omitempty"`
	DurationSeconds int                        `json:"duration_seconds"`
	AspectRatio     string                     `json:"aspect_ratio,omitempty"`
	Title           string                     `json:"title"`
	Summary         string                     `json:"summary"`
	ScriptText      string                     `json:"script_text"`
	NarrationText   string                     `json:"narration_text"`
	PlannerProvider string                     `json:"planner_provider,omitempty"`
	PlannerModel    string                     `json:"planner_model,omitempty"`
	ImageProvider   string                     `json:"image_provider,omitempty"`
	ImageModel      string                     `json:"image_model,omitempty"`
	AudioProvider   string                     `json:"audio_provider,omitempty"`
	AudioModel      string                     `json:"audio_model,omitempty"`
	AudioVoice      string                     `json:"audio_voice,omitempty"`
	Shots           []StoryVideoDraftShotInput `json:"shots"`
}

type StoryVideoShot struct {
	ID            string `json:"id"`
	Index         int    `json:"index"`
	Title         string `json:"title"`
	StoryBeat     string `json:"story_beat"`
	NarrationLine string `json:"narration_line"`
	ImagePrompt   string `json:"image_prompt"`
	ImageURL      string `json:"image_url,omitempty"`
	AudioURL      string `json:"audio_url,omitempty"`
	DurationMS    int    `json:"duration_ms"`
	Status        string `json:"status"`
	AttemptCount  int    `json:"attempt_count"`
	Error         string `json:"error,omitempty"`
}

type StoryVideoProject struct {
	ID              string           `json:"id"`
	Status          string           `json:"status"`
	Keywords        []string         `json:"keywords"`
	Theme           string           `json:"theme,omitempty"`
	Audience        string           `json:"audience,omitempty"`
	Tone            string           `json:"tone,omitempty"`
	Extra           string           `json:"extra,omitempty"`
	DurationSeconds int              `json:"duration_seconds"`
	AspectRatio     string           `json:"aspect_ratio,omitempty"`
	Title           string           `json:"title,omitempty"`
	Summary         string           `json:"summary,omitempty"`
	ScriptText      string           `json:"script_text,omitempty"`
	NarrationText   string           `json:"narration_text,omitempty"`
	PlannerProvider string           `json:"planner_provider,omitempty"`
	PlannerModel    string           `json:"planner_model,omitempty"`
	ImageProvider   string           `json:"image_provider,omitempty"`
	ImageModel      string           `json:"image_model,omitempty"`
	AudioProvider   string           `json:"audio_provider,omitempty"`
	AudioModel      string           `json:"audio_model,omitempty"`
	AudioVoice      string           `json:"audio_voice,omitempty"`
	AudioURL        string           `json:"audio_url,omitempty"`
	VideoURL        string           `json:"video_url,omitempty"`
	Error           string           `json:"error,omitempty"`
	CreatedAt       string           `json:"created_at,omitempty"`
	UpdatedAt       string           `json:"updated_at,omitempty"`
	Shots           []StoryVideoShot `json:"shots,omitempty"`
}

type StoryVideoEvent struct {
	ID        uint64 `json:"id"`
	Stage     string `json:"stage"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Payload   string `json:"payload,omitempty"`
	CreatedAt string `json:"created_at"`
}

type StoryVideoRegenerateAudioRequest struct {
	NarrationText string `json:"narration_text,omitempty"`
	AudioProvider string `json:"audio_provider,omitempty"`
	AudioModel    string `json:"audio_model,omitempty"`
	AudioVoice    string `json:"audio_voice,omitempty"`
}

type StoryVideoRegenerateShotRequest struct {
	ImagePrompt   string `json:"image_prompt,omitempty"`
	ImageProvider string `json:"image_provider,omitempty"`
	ImageModel    string `json:"image_model,omitempty"`
}
