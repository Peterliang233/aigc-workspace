package storyvideo

import (
	"fmt"
	"strings"
)

func repairDraft(in Draft, req DraftRequest) Draft {
	out := in
	if strings.TrimSpace(out.Title) == "" {
		out.Title = fallbackTitle(req)
	}
	if strings.TrimSpace(out.Summary) == "" {
		out.Summary = fallbackSummary(out, req)
	}
	if strings.TrimSpace(out.ScriptText) == "" {
		out.ScriptText = fallbackScript(out)
	}
	if strings.TrimSpace(out.NarrationText) == "" {
		out.NarrationText = fallbackNarration(out)
	}
	for i := range out.Shots {
		shot := &out.Shots[i]
		if strings.TrimSpace(shot.Title) == "" {
			shot.Title = fmt.Sprintf("分镜 %d", i+1)
		}
		if strings.TrimSpace(shot.StoryBeat) == "" {
			shot.StoryBeat = firstNonEmptyText(shot.NarrationLine, shot.Title, out.Summary)
		}
		if strings.TrimSpace(shot.NarrationLine) == "" {
			shot.NarrationLine = firstNonEmptyText(shot.StoryBeat, shot.Title, fmt.Sprintf("%s 的关键一幕。", shot.Title))
		}
		if strings.TrimSpace(shot.ImagePrompt) == "" {
			shot.ImagePrompt = strings.TrimSpace(strings.Join([]string{
				shot.Title,
				shot.StoryBeat,
				shot.NarrationLine,
				strings.TrimSpace(req.Theme),
				strings.TrimSpace(req.Tone),
			}, "，"))
		}
	}
	return out
}

func hasDraftSignal(in Draft) bool {
	if strings.TrimSpace(in.Title) != "" || strings.TrimSpace(in.Summary) != "" ||
		strings.TrimSpace(in.ScriptText) != "" || strings.TrimSpace(in.NarrationText) != "" {
		return true
	}
	for _, shot := range in.Shots {
		if strings.TrimSpace(shot.Title) != "" || strings.TrimSpace(shot.StoryBeat) != "" ||
			strings.TrimSpace(shot.NarrationLine) != "" || strings.TrimSpace(shot.ImagePrompt) != "" {
			return true
		}
	}
	return false
}

func fallbackTitle(req DraftRequest) string {
	if len(req.Keywords) == 0 {
		return "故事短片"
	}
	return strings.Join(req.Keywords, "·") + "短片"
}

func fallbackSummary(in Draft, req DraftRequest) string {
	return firstNonEmptyText(in.ScriptText, in.NarrationText, strings.Join(req.Keywords, "、")+" 的故事演绎。")
}

func fallbackScript(in Draft) string {
	lines := make([]string, 0, len(in.Shots))
	for _, shot := range in.Shots {
		lines = append(lines, firstNonEmptyText(shot.StoryBeat, shot.NarrationLine, shot.Title))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func fallbackNarration(in Draft) string {
	lines := make([]string, 0, len(in.Shots))
	for _, shot := range in.Shots {
		lines = append(lines, firstNonEmptyText(shot.NarrationLine, shot.StoryBeat, shot.Title))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
