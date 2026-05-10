import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";
import { useStoryVideo } from "../../state/storyvideo";

export function StoryVideoVideoPanel(props: { project: StoryVideoProject; busy: boolean; error: string }) {
  const { composeProject } = useStoryVideo();
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">最终成片</h2>
        <button className="btn" disabled={props.busy} onClick={() => void composeProject()}>合成视频</button>
      </div>
      {props.error ? <div className="alert alert--err">{props.error}</div> : null}
      <div className="panel"><div className="k">完成条件</div><div>所有分镜图片与解说音频准备完成后即可合成最终视频。</div></div>
      {props.project.video_url ? <video className="storyvideo__video" controls src={props.project.video_url} /> : <div className="storyvideo__emptyMedia">等待音频和全部分镜完成后合成。</div>}
    </section>
  );
}
