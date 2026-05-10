package storyvideo

import "time"

type Project struct {
	ID                string    `gorm:"column:id;primaryKey;size:64"`
	Status            string    `gorm:"column:status;size:32;index"`
	KeywordsJSON      string    `gorm:"column:keywords_json;type:json"`
	Theme             string    `gorm:"column:theme;size:128"`
	Audience          string    `gorm:"column:audience;size:128"`
	Tone              string    `gorm:"column:tone;size:128"`
	ExtraRequirements string    `gorm:"column:extra_requirements;type:text"`
	DurationSeconds   int       `gorm:"column:duration_seconds"`
	AspectRatio       string    `gorm:"column:aspect_ratio;size:32"`
	Title             string    `gorm:"column:title;size:255"`
	Summary           string    `gorm:"column:summary;type:text"`
	ScriptText        string    `gorm:"column:script_text;type:longtext"`
	NarrationText     string    `gorm:"column:narration_text;type:longtext"`
	PlannerProvider   string    `gorm:"column:planner_provider;size:64"`
	PlannerModel      string    `gorm:"column:planner_model;size:128"`
	ImageProvider     string    `gorm:"column:image_provider;size:64"`
	ImageModel        string    `gorm:"column:image_model;size:255"`
	AudioProvider     string    `gorm:"column:audio_provider;size:64"`
	AudioModel        string    `gorm:"column:audio_model;size:255"`
	AudioVoice        string    `gorm:"column:audio_voice;size:128"`
	AudioAssetID      *uint64   `gorm:"column:audio_asset_id"`
	VideoAssetID      *uint64   `gorm:"column:video_asset_id"`
	Error             *string   `gorm:"column:error;type:text"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Project) TableName() string { return "aigc_story_video_projects" }

type Shot struct {
	ID            string    `gorm:"column:id;primaryKey;size:64"`
	ProjectID     string    `gorm:"column:project_id;size:64;index"`
	ShotIndex     int       `gorm:"column:shot_index"`
	Title         string    `gorm:"column:title;size:255"`
	StoryBeat     string    `gorm:"column:story_beat;type:text"`
	NarrationLine string    `gorm:"column:narration_line;type:text"`
	ImagePrompt   string    `gorm:"column:image_prompt;type:text"`
	ImageAssetID  *uint64   `gorm:"column:image_asset_id"`
	AudioAssetID  *uint64   `gorm:"column:audio_asset_id"`
	DurationMS    int       `gorm:"column:duration_ms"`
	Status        string    `gorm:"column:status;size:32;index"`
	AttemptCount  int       `gorm:"column:attempt_count"`
	Error         *string   `gorm:"column:error;type:text"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Shot) TableName() string { return "aigc_story_video_shots" }

type Event struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	ProjectID string    `gorm:"column:project_id;size:64;index"`
	Stage     string    `gorm:"column:stage;size:32;index"`
	Type      string    `gorm:"column:event_type;size:64"`
	Message   string    `gorm:"column:message;type:text"`
	Payload   string    `gorm:"column:payload_json;type:json"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

func (Event) TableName() string { return "aigc_story_video_events" }
