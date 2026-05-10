import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";
import { useStoryVideo } from "../../state/storyvideo";

export function StoryVideoAudioPanel(props: { project: StoryVideoProject; busy: boolean; error: string }) {
  const { regenerateAudio } = useStoryVideo();
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">解说音频</h2>
        <button className="btn btn--ghost" disabled={props.busy} onClick={() => void regenerateAudio({ narration_text: props.project.narration_text })}>重生成音频</button>
      </div>
      {props.error ? <div className="alert alert--err">{props.error}</div> : null}
      {props.project.error ? <div className="alert alert--err">{props.project.error}</div> : null}
      <div className="panel"><div className="k">解说文本</div><div>{props.project.narration_text || "暂无解说词"}</div></div>
      <div className="panel"><div className="k">音频模型</div><div>{props.project.audio_provider} · {props.project.audio_model || "-"}</div></div>
      {props.project.audio_url ? <audio className="storyvideo__audio" controls src={props.project.audio_url} /> : <div className="panel"><div className="k">尚未生成音频</div></div>}
      <div className="storyvideoShots">
        {(props.project.shots || []).map((shot) => (
          <div key={shot.id} className="panel">
            <div className="panel__row"><strong>{shot.index}. {shot.title || "未命名分镜"}</strong><span className="k">{shot.duration_ms}ms</span></div>
            <div className="k">{shot.narration_line || "暂无对应台词"}</div>
            {shot.audio_url ? <audio className="storyvideo__audio" controls src={shot.audio_url} /> : <div className="storyvideo__emptyMedia">分镜音频未生成</div>}
          </div>
        ))}
      </div>
    </section>
  );
}
