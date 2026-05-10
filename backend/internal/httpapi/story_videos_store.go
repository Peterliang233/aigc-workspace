package httpapi

import (
	"context"
	"fmt"
	"strings"

	"aigc-backend/internal/storyvideo"
	"aigc-backend/internal/types"
)

func (h *Handler) storyVideoCreateDraft(ctx context.Context, req types.StoryVideoDraftRequest, draft storyvideo.Draft) (types.StoryVideoProject, error) {
	project := &storyvideo.Project{
		ID:                storyVideoNewID("sv"),
		Status:            "draft_ready",
		KeywordsJSON:      storyVideoKeywordsJSON(req.Keywords),
		Theme:             req.Theme,
		Audience:          req.Audience,
		Tone:              req.Tone,
		ExtraRequirements: req.Extra,
		DurationSeconds:   req.DurationSeconds,
		AspectRatio:       req.AspectRatio,
		Title:             strings.TrimSpace(draft.Title),
		Summary:           strings.TrimSpace(draft.Summary),
		ScriptText:        strings.TrimSpace(draft.ScriptText),
		NarrationText:     strings.TrimSpace(draft.NarrationText),
		PlannerProvider:   req.PlannerProvider,
		PlannerModel:      req.PlannerModel,
		ImageProvider:     req.ImageProvider,
		ImageModel:        req.ImageModel,
		AudioProvider:     req.AudioProvider,
		AudioModel:        req.AudioModel,
	}
	shots := make([]storyvideo.Shot, 0, len(draft.Shots))
	for i, shot := range draft.Shots {
		shots = append(shots, storyvideo.Shot{
			ID:            storyVideoNewID("shot"),
			ProjectID:     project.ID,
			ShotIndex:     i + 1,
			Title:         strings.TrimSpace(shot.Title),
			StoryBeat:     strings.TrimSpace(shot.StoryBeat),
			NarrationLine: strings.TrimSpace(shot.NarrationLine),
			ImagePrompt:   strings.TrimSpace(shot.ImagePrompt),
			DurationMS:    shot.DurationMS,
			Status:        "draft",
		})
	}
	if len(shots) == 0 {
		return types.StoryVideoProject{}, fmt.Errorf("规划结果没有分镜")
	}
	if err := h.storyVideos.CreateProject(ctx, project, shots); err != nil {
		return types.StoryVideoProject{}, err
	}
	_ = h.storyVideoAddEvent(ctx, project.ID, "draft", "draft_ready", "故事台本已生成", map[string]any{"shot_count": len(shots)})
	return h.storyVideoResponse(project, shots), nil
}

func (h *Handler) storyVideoReplaceDraft(ctx context.Context, projectID string, req types.StoryVideoDraftUpdateRequest) (types.StoryVideoProject, error) {
	project, _, err := h.storyVideos.GetProject(ctx, projectID)
	if err != nil {
		return types.StoryVideoProject{}, err
	}
	if strings.TrimSpace(project.Status) != "draft_ready" {
		return types.StoryVideoProject{}, fmt.Errorf("草稿已确认，不能继续保存草稿")
	}
	req.Keywords = storyVideoCleanList(req.Keywords)
	if len(req.Keywords) == 0 || len(req.Shots) == 0 {
		return types.StoryVideoProject{}, fmt.Errorf("关键词和分镜不能为空")
	}
	project.Status = "draft_ready"
	project.KeywordsJSON = storyVideoKeywordsJSON(req.Keywords)
	project.Theme = strings.TrimSpace(req.Theme)
	project.Audience = strings.TrimSpace(req.Audience)
	project.Tone = strings.TrimSpace(req.Tone)
	project.ExtraRequirements = strings.TrimSpace(req.Extra)
	project.DurationSeconds = req.DurationSeconds
	project.AspectRatio = strings.TrimSpace(req.AspectRatio)
	project.Title = strings.TrimSpace(req.Title)
	project.Summary = strings.TrimSpace(req.Summary)
	project.ScriptText = strings.TrimSpace(req.ScriptText)
	project.NarrationText = strings.TrimSpace(req.NarrationText)
	project.AudioAssetID = nil
	project.VideoAssetID = nil
	project.Error = nil
	project.PlannerProvider = firstNonEmpty(req.PlannerProvider, project.PlannerProvider)
	project.PlannerModel = firstNonEmpty(req.PlannerModel, project.PlannerModel)
	project.ImageProvider = firstNonEmpty(req.ImageProvider, project.ImageProvider)
	project.ImageModel = firstNonEmpty(req.ImageModel, project.ImageModel)
	project.AudioProvider = firstNonEmpty(req.AudioProvider, project.AudioProvider)
	project.AudioModel = firstNonEmpty(req.AudioModel, project.AudioModel)
	shots := make([]storyvideo.Shot, 0, len(req.Shots))
	for i, shot := range req.Shots {
		shots = append(shots, storyvideo.Shot{
			ID:            firstNonEmpty(shot.ID, storyVideoNewID("shot")),
			ProjectID:     projectID,
			ShotIndex:     i + 1,
			Title:         strings.TrimSpace(shot.Title),
			StoryBeat:     strings.TrimSpace(shot.StoryBeat),
			NarrationLine: strings.TrimSpace(shot.NarrationLine),
			ImagePrompt:   strings.TrimSpace(shot.ImagePrompt),
			DurationMS:    maxInt(1000, shot.DurationMS),
			Status:        "draft",
		})
	}
	if err := h.storyVideos.ReplaceDraft(ctx, project, shots); err != nil {
		return types.StoryVideoProject{}, err
	}
	_ = h.storyVideoAddEvent(ctx, projectID, "draft", "draft_updated", "草稿已保存", map[string]any{"shot_count": len(shots)})
	return h.storyVideoResponse(project, shots), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
