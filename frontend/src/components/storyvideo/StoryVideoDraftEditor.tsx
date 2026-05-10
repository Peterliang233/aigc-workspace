import React, { useEffect, useState } from "react";
import type { StoryVideoDraftUpdateRequest, StoryVideoProject } from "../../api_storyvideo";
import { useStoryVideo } from "../../state/storyvideo";

export function StoryVideoDraftEditor(props: { project: StoryVideoProject | null; busy: boolean }) {
  const { saveDraft, confirmProject } = useStoryVideo();
  const [draft, setDraft] = useState<StoryVideoDraftUpdateRequest | null>(null);
  useEffect(() => {
    if (!props.project) return setDraft(null);
    setDraft({
      keywords: props.project.keywords || [],
      theme: props.project.theme,
      audience: props.project.audience,
      tone: props.project.tone,
      extra: props.project.extra,
      duration_seconds: props.project.duration_seconds,
      aspect_ratio: props.project.aspect_ratio,
      title: props.project.title || "",
      summary: props.project.summary || "",
      script_text: props.project.script_text || "",
      narration_text: props.project.narration_text || "",
      planner_provider: props.project.planner_provider,
      planner_model: props.project.planner_model,
      image_provider: props.project.image_provider,
      image_model: props.project.image_model,
      audio_provider: props.project.audio_provider,
      audio_model: props.project.audio_model,
      audio_voice: props.project.audio_voice,
      shots: (props.project.shots || []).map((shot) => ({ id: shot.id, title: shot.title, story_beat: shot.story_beat, narration_line: shot.narration_line, image_prompt: shot.image_prompt, duration_ms: shot.duration_ms }))
    });
  }, [props.project]);
  if (!props.project || !draft) return null;
  const canEditDraft = props.project.status === "draft_ready";
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">草稿确认</h2>
        <span className="badge">{props.project.status}</span>
      </div>
      <div className="form">
        {!canEditDraft ? <div className="panel"><div className="k">草稿已确认</div><div>素材生成流程已接管，草稿不会再重复提交。</div></div> : null}
        <fieldset className="storyvideoDraft__fields" disabled={!canEditDraft}>
          <label className="label">标题<input className="input" value={draft.title} onChange={(e) => setDraft({ ...draft, title: e.target.value })} /></label>
          <label className="label">摘要<textarea className="textarea" rows={3} value={draft.summary} onChange={(e) => setDraft({ ...draft, summary: e.target.value })} /></label>
          <label className="label">故事台本<textarea className="textarea" rows={5} value={draft.script_text} onChange={(e) => setDraft({ ...draft, script_text: e.target.value })} /></label>
          <label className="label">解说词<textarea className="textarea" rows={5} value={draft.narration_text} onChange={(e) => setDraft({ ...draft, narration_text: e.target.value })} /></label>
          <div className="storyvideo__shots">
            {draft.shots.map((shot, index) => (
              <div key={shot.id || index} className="panel">
                <div className="panel__row"><strong>分镜 {index + 1}</strong><span className="k">{shot.duration_ms}ms</span></div>
                <input className="input" value={shot.title} onChange={(e) => setDraft({ ...draft, shots: draft.shots.map((item, idx) => idx === index ? { ...item, title: e.target.value } : item) })} />
                <textarea className="textarea" rows={2} value={shot.narration_line} onChange={(e) => setDraft({ ...draft, shots: draft.shots.map((item, idx) => idx === index ? { ...item, narration_line: e.target.value } : item) })} />
                <textarea className="textarea" rows={3} value={shot.image_prompt} onChange={(e) => setDraft({ ...draft, shots: draft.shots.map((item, idx) => idx === index ? { ...item, image_prompt: e.target.value } : item) })} />
              </div>
            ))}
          </div>
        </fieldset>
        {canEditDraft ? <div className="storyvideo__actions">
          <button className="btn btn--ghost" disabled={props.busy} onClick={() => void saveDraft(draft)}>保存草稿</button>
          <button className="btn" disabled={props.busy} onClick={() => void confirmProject()}>确认并生成素材</button>
        </div> : null}
      </div>
    </section>
  );
}
