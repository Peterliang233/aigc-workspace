package assets

import "time"

type Asset struct {
	ID uint64 `gorm:"column:id;primaryKey;autoIncrement"`

	Capability string `gorm:"column:capability;size:16;not null;index:idx_cap_created,priority:1"`
	Provider   string `gorm:"column:provider;size:64;not null"`
	Model      string `gorm:"column:model;size:255;not null;default:''"`

	PromptSHA256  string `gorm:"column:prompt_sha256;size:64;not null;default:''"`
	PromptPreview string `gorm:"column:prompt_preview;size:255;not null;default:''"`

	ParamsJSON string `gorm:"column:params_json;type:json"`

	Status string  `gorm:"column:status;size:16;not null;default:'succeeded'"`
	Error  *string `gorm:"column:error;type:text"`

	SourceURL *string `gorm:"column:source_url;type:text"`

	ObjectKey     string  `gorm:"column:object_key;size:512;not null"`
	ContentType   string  `gorm:"column:content_type;size:128;not null;default:'application/octet-stream'"`
	Bytes         int64   `gorm:"column:bytes;not null;default:0"`
	ExternalJobID *string `gorm:"column:external_job_id;size:128;uniqueIndex"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_cap_created,priority:2"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Asset) TableName() string { return "aigc_generation_assets" }
