package assets

import (
	"strings"

	"aigc-backend/internal/blobstore"
	"aigc-backend/internal/util/mediafetch"
)

type Service struct {
	Store    Store
	MinIO    *blobstore.MinIO
	Download *mediafetch.Downloader

	// MaxBytes caps the size of a single fetched asset to avoid runaway downloads.
	MaxBytes int64
}

type StoreRemoteInput struct {
	Capability    string
	Provider      string
	Model         string
	Prompt        string
	Params        any
	SourceURL     string
	ExternalJobID string // optional (video job id)
}

type StoreLocalFileInput struct {
	Capability string
	Provider   string
	Model      string
	Prompt     string
	Params     any
	FilePath   string
	// ContentType is optional; if empty, it will be sniffed from the file.
	ContentType string
}

func (s *Service) Enabled() bool {
	return s != nil && s.Store != nil && s.MinIO != nil && s.MinIO.Client != nil && strings.TrimSpace(s.MinIO.Bucket) != ""
}

