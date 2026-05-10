import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";
import { useStoryVideo } from "../../state/storyvideo";

export function StoryVideoShotsPanel(props: { project: StoryVideoProject; busy: boolean }) {
  const { regenerateShot } = useStoryVideo();
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">分镜图片</h2>
        <span className="badge">{(props.project.shots || []).length} 个分镜</span>
      </div>
      <div className="storyvideoShots">
        {(props.project.shots || []).map((shot) => (
          <div key={shot.id} className="panel">
            <div className="panel__row"><strong>{shot.index}. {shot.title || "未命名分镜"}</strong><span className="pill">{shot.status}</span></div>
            <div className="k">{shot.story_beat || "暂无剧情描述"}</div>
            <div className="k">{shot.narration_line || "暂无对应台词"}</div>
            <div className="k">时长：{shot.duration_ms}ms · 重试：{shot.attempt_count}</div>
            {shot.image_url ? <img className="storyvideo__image" src={shot.image_url} alt={shot.title || shot.id} /> : <div className="storyvideo__emptyMedia">图片未生成</div>}
            {shot.error ? <div className="alert alert--err">{shot.error}</div> : null}
            <button className="btn btn--ghost" disabled={props.busy} onClick={() => void regenerateShot(shot.id, { image_prompt: shot.image_prompt })}>重生成图片</button>
          </div>
        ))}
      </div>
    </section>
  );
}
