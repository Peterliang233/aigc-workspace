import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";
import { useStoryVideo } from "../../state/storyvideo";

export function StoryVideoVideoPanel(props: { project: StoryVideoProject; busy: boolean; error: string }) {
  const { composeProject } = useStoryVideo();
  const hasVideo = !!props.project.video_url || props.project.status === "succeeded";
  const isComposing = props.project.status === "composing";
  const shotsReady = (props.project.shots || []).length > 0 && (props.project.shots || []).every((shot) => !!shot.image_url || shot.status === "succeeded");
  const canCompose = !hasVideo && !isComposing && !!props.project.audio_url && shotsReady;
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">最终成片</h2>
        {!hasVideo ? <button className="btn" disabled={props.busy || !canCompose} onClick={() => void composeProject()}>{isComposing ? "合成中..." : "合成视频"}</button> : null}
      </div>
      {props.error ? <div className="alert alert--err">{props.error}</div> : null}
      <div className="panel"><div className="k">完成条件</div><div>所有分镜图片与解说音频准备完成后即可合成最终视频。</div></div>
      {props.project.video_url ? <video className="storyvideo__video" controls src={props.project.video_url} /> : <div className="storyvideo__emptyMedia">等待音频和全部分镜完成后合成。</div>}
    </section>
  );
}
