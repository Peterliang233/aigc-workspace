import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";
import { useStoryVideo } from "../../state/storyvideo";

export function StoryVideoProgress(props: { project: StoryVideoProject | null; busy: boolean; error: string }) {
  const { regenerateAudio, regenerateShot, composeProject } = useStoryVideo();
  if (!props.project) return null;
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">生成进度</h2>
        <span className="badge">{props.project.status}</span>
      </div>
      {props.error ? <div className="alert alert--err">{props.error}</div> : null}
      {props.project.error ? <div className="alert alert--err">{props.project.error}</div> : null}
      <div className="panel">
        <div className="panel__row"><span className="k">解说音频</span><button className="btn btn--ghost" disabled={props.busy} onClick={() => void regenerateAudio({ narration_text: props.project?.narration_text })}>重生成音频</button></div>
        {props.project.audio_url ? <audio className="storyvideo__audio" controls src={props.project.audio_url} /> : <div className="k">尚未生成</div>}
      </div>
      <div className="storyvideo__shots">
        {(props.project.shots || []).map((shot) => (
          <div key={shot.id} className="panel">
            <div className="panel__row"><strong>{shot.index}. {shot.title}</strong><span className="pill">{shot.status}</span></div>
            <div className="k">{shot.narration_line}</div>
            <div className="k">尝试次数：{shot.attempt_count}</div>
            {shot.image_url ? <img className="storyvideo__image" src={shot.image_url} alt={shot.title} /> : <div className="k">图片未生成</div>}
            {shot.error ? <div className="alert alert--err">{shot.error}</div> : null}
            <button className="btn btn--ghost" disabled={props.busy} onClick={() => void regenerateShot(shot.id, { image_prompt: shot.image_prompt })}>重生成图片</button>
          </div>
        ))}
      </div>
      <div className="panel">
        <div className="panel__row"><span className="k">最终视频</span><button className="btn" disabled={props.busy} onClick={() => void composeProject()}>合成视频</button></div>
        {props.project.video_url ? <video className="storyvideo__video" controls src={props.project.video_url} /> : <div className="k">等待音频和全部分镜完成后合成。</div>}
      </div>
    </section>
  );
}
