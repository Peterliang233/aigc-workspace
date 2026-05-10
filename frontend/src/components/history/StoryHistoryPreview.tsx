import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";

export function StoryHistoryPreview(props: { project: StoryVideoProject }) {
  const project = props.project;
  return (
    <>
      <div className="panel">
        <div className="panel__row"><div className="k">标题</div><div className="v">{project.title || "-"}</div></div>
        <div className="panel__row"><div className="k">关键词</div><div className="v">{(project.keywords || []).join(" / ") || "-"}</div></div>
        <div className="panel__row"><div className="k">图片模型</div><div className="v">{project.image_provider} · {project.image_model || "-"}</div></div>
        <div className="panel__row"><div className="k">音频模型</div><div className="v">{project.audio_provider} · {project.audio_model || "-"}</div></div>
      </div>
      {project.summary ? <div className="panel"><div className="k">摘要</div><div>{project.summary}</div></div> : null}
      {project.video_url ? <div className="panel__media"><video className="video" controls src={project.video_url} /></div> : null}
      {project.audio_url ? <div className="panel__media"><audio className="audioPlayer" controls src={project.audio_url} /></div> : null}
      <div className="list">
        {(project.shots || []).map((shot) => (
          <div key={shot.id} className="panel">
            <div className="panel__row"><div className="k">{shot.index}. {shot.title || "未命名分镜"}</div><div className="v">{shot.status}</div></div>
            {shot.image_url ? <img className="preview__img" src={shot.image_url} alt={shot.title || shot.id} /> : null}
            {shot.narration_line ? <div className="k">{shot.narration_line}</div> : null}
          </div>
        ))}
      </div>
    </>
  );
}
