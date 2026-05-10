package httpapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aigc-backend/internal/assets"
	"aigc-backend/internal/storyvideo"
	"aigc-backend/internal/types"
)

type storyVideoAudioSegment struct {
	ShotID string
	Text   string
}

func (h *Handler) runStoryVideoAudioSync(ctx context.Context, projectID string) error {
	project, shots, err := h.storyVideos.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	if h.storyMedia == nil || !h.storyMedia.Enabled() {
		return h.storyVideoFailProject(ctx, projectID, "audio", fmt.Errorf("媒体合成服务未配置"))
	}
	if h.assets == nil || !h.assets.Enabled() {
		return h.storyVideoFailProject(ctx, projectID, "audio", fmt.Errorf("素材存储未配置"))
	}
	segments := storyVideoNarrationSegments(project.NarrationText, shots)
	narrationText := storyVideoSegmentsText(segments)
	if strings.TrimSpace(narrationText) == "" {
		return h.storyVideoFailProject(ctx, projectID, "audio", fmt.Errorf("解说词不能为空"))
	}
	_ = h.storyVideoAddEvent(ctx, projectID, "audio", "segments_planned", fmt.Sprintf("已拆分 %d 段分镜解说，开始逐段生成音频", len(segments)), map[string]any{"segments": len(segments)})
	return h.storyVideoGenerateAudioSegments(ctx, project, segments, narrationText)
}

func (h *Handler) storyVideoGenerateAudioSegments(ctx context.Context, project *storyvideo.Project, segments []storyVideoAudioSegment, narrationText string) error {
	providerID := firstNonEmpty(project.AudioProvider, h.defaultAudioProviderID())
	defaultModel := ""
	if h.models != nil {
		defaultModel = h.models.DefaultModel(providerID, "audio")
	}
	model := firstNonEmpty(project.AudioModel, defaultModel)
	prov, ok := h.getAudioProvider(providerID)
	if !ok || prov == nil {
		return h.storyVideoFailProject(ctx, project.ID, "audio", fmt.Errorf("音频 provider 不可用"))
	}
	tmpDir, err := os.MkdirTemp("", "story-video-audio-*")
	if err != nil {
		return h.storyVideoFailProject(ctx, project.ID, "audio", err)
	}
	defer os.RemoveAll(tmpDir)
	_ = h.storyVideoClearShotAudioAssets(ctx, segments)
	paths, err := h.storyVideoGenerateSegmentFiles(ctx, prov, providerID, model, project.AudioVoice, segments, tmpDir)
	if err != nil {
		return h.storyVideoFailProject(ctx, project.ID, "audio", err)
	}
	mergedPath := filepath.Join(tmpDir, "narration.m4a")
	_ = h.storyVideoAddEvent(ctx, project.ID, "audio", "segments_generated", "分镜音频片段已生成，开始拼接完整解说音频", map[string]any{"segments": len(paths)})
	durations, err := h.storyMedia.ConcatAudios(ctx, paths, mergedPath)
	if err != nil {
		return h.storyVideoFailProject(ctx, project.ID, "audio", err)
	}
	_ = h.storyVideoUpdateShotDurations(ctx, segments, durations)
	asset, err := h.assets.StoreLocalFile(ctx, assets.StoreLocalFileInput{
		Capability: "audio", Provider: providerID, Model: model, Prompt: narrationText,
		Params: map[string]any{"voice": project.AudioVoice, "segments": len(segments)}, FilePath: mergedPath, ContentType: "audio/mp4",
	})
	if err != nil {
		return h.storyVideoFailProject(ctx, project.ID, "audio", err)
	}
	_ = h.storyVideos.UpdateProject(ctx, project.ID, map[string]any{"narration_text": narrationText, "audio_asset_id": asset.ID, "error": nil})
	_ = h.storyVideoAddEvent(ctx, project.ID, "audio", "succeeded", "分镜解说音频拼接成功", map[string]any{"asset_id": asset.ID, "segments": len(segments)})
	return nil
}

func (h *Handler) storyVideoGenerateSegmentFiles(ctx context.Context, prov audioProvider, providerID, model, voice string, segments []storyVideoAudioSegment, tmpDir string) ([]string, error) {
	paths := make([]string, 0, len(segments))
	for i, segment := range segments {
		req := types.AudioGenerateRequest{Provider: providerID, Model: model, Input: segment.Text, Voice: voice}
		h.applyAudioModelDefaults(providerID, model, &req)
		if miss := h.missingAudioRequiredFields(providerID, model, req); len(miss) > 0 {
			return nil, fmt.Errorf("缺少音频参数: %s", strings.Join(miss, ", "))
		}
		var resp types.AudioGenerateResponse
		err := retryDownstreamCall(ctx, "story_video_audio_generate", func(callCtx context.Context) error {
			var callErr error
			resp, callErr = prov.GenerateAudio(callCtx, req)
			return callErr
		})
		if err != nil {
			return nil, err
		}
		path, err := h.storyVideoMaterializeAudio(ctx, tmpDir, i+1, resp)
		if err != nil {
			return nil, err
		}
		if err := h.storyVideoStoreShotAudio(ctx, segment, path, providerID, model, req); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func storyVideoNarrationForAudio(projectNarration string, shots []storyvideo.Shot) string {
	return storyVideoSegmentsText(storyVideoNarrationSegments(projectNarration, shots))
}

func storyVideoNarrationSegments(projectNarration string, shots []storyvideo.Shot) []storyVideoAudioSegment {
	out := make([]storyVideoAudioSegment, 0, len(shots))
	hasLine := false
	for _, shot := range shots {
		if strings.TrimSpace(shot.NarrationLine) != "" {
			hasLine = true
			break
		}
	}
	if !hasLine {
		return []storyVideoAudioSegment{{Text: strings.TrimSpace(projectNarration)}}
	}
	for _, shot := range shots {
		text := firstNonEmpty(shot.NarrationLine, shot.StoryBeat, shot.Title)
		if text == "" {
			continue
		}
		out = append(out, storyVideoAudioSegment{ShotID: shot.ID, Text: text})
	}
	return out
}

func storyVideoSegmentsText(segments []storyVideoAudioSegment) string {
	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	return strings.Join(lines, "\n")
}

func (h *Handler) storyVideoUpdateShotDurations(ctx context.Context, segments []storyVideoAudioSegment, durations []int) error {
	for i, duration := range durations {
		if i >= len(segments) || strings.TrimSpace(segments[i].ShotID) == "" {
			continue
		}
		if err := h.storyVideos.UpdateShot(ctx, segments[i].ShotID, map[string]any{"duration_ms": maxInt(1000, duration)}); err != nil {
			return err
		}
	}
	return nil
}
