package httpapi

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"aigc-backend/internal/storyvideo"
	"aigc-backend/internal/types"
)

var storyVideoIDSeq uint64

func (h *Handler) storyVideoReady() error {
	if h.storyVideos == nil {
		return fmt.Errorf("故事视频能力依赖 MySQL 持久化，请先配置数据库")
	}
	return nil
}

func storyVideoProjectID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 || parts[0] != "api" || parts[1] != "story-videos" || parts[2] != "projects" {
		return ""
	}
	return strings.TrimSpace(parts[3])
}

func storyVideoShotID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 7 || parts[4] != "shots" {
		return ""
	}
	return strings.TrimSpace(parts[5])
}

func storyVideoAssetURL(id *uint64) string {
	if id == nil || *id == 0 {
		return ""
	}
	return fmt.Sprintf("/api/assets/%d", *id)
}

func storyVideoNewID(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), atomic.AddUint64(&storyVideoIDSeq, 1))
}

func storyVideoKeywords(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func storyVideoErr(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(*raw)
}

func (h *Handler) storyVideoResponse(project *storyvideo.Project, shots []storyvideo.Shot) types.StoryVideoProject {
	out := types.StoryVideoProject{
		ID:              project.ID,
		Status:          project.Status,
		Keywords:        storyVideoKeywords(project.KeywordsJSON),
		Theme:           project.Theme,
		Audience:        project.Audience,
		Tone:            project.Tone,
		Extra:           project.ExtraRequirements,
		DurationSeconds: project.DurationSeconds,
		AspectRatio:     project.AspectRatio,
		Title:           project.Title,
		Summary:         project.Summary,
		ScriptText:      project.ScriptText,
		NarrationText:   project.NarrationText,
		PlannerProvider: project.PlannerProvider,
		PlannerModel:    project.PlannerModel,
		ImageProvider:   project.ImageProvider,
		ImageModel:      project.ImageModel,
		AudioProvider:   project.AudioProvider,
		AudioModel:      project.AudioModel,
		AudioVoice:      project.AudioVoice,
		AudioURL:        storyVideoAssetURL(project.AudioAssetID),
		VideoURL:        storyVideoAssetURL(project.VideoAssetID),
		Error:           storyVideoErr(project.Error),
		CreatedAt:       project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       project.UpdatedAt.Format(time.RFC3339),
	}
	out.Shots = make([]types.StoryVideoShot, 0, len(shots))
	for _, shot := range shots {
		out.Shots = append(out.Shots, types.StoryVideoShot{
			ID:            shot.ID,
			Index:         shot.ShotIndex,
			Title:         shot.Title,
			StoryBeat:     shot.StoryBeat,
			NarrationLine: shot.NarrationLine,
			ImagePrompt:   shot.ImagePrompt,
			ImageURL:      storyVideoAssetURL(shot.ImageAssetID),
			AudioURL:      storyVideoAssetURL(shot.AudioAssetID),
			DurationMS:    shot.DurationMS,
			Status:        shot.Status,
			AttemptCount:  shot.AttemptCount,
			Error:         storyVideoErr(shot.Error),
		})
	}
	return out
}

func storyVideoEventResponse(event storyvideo.Event) types.StoryVideoEvent {
	return types.StoryVideoEvent{
		ID:        event.ID,
		Stage:     event.Stage,
		Type:      event.Type,
		Message:   event.Message,
		Payload:   event.Payload,
		CreatedAt: event.CreatedAt.Format(time.RFC3339),
	}
}
