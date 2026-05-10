import React, { useEffect, useState } from "react";
import type { StoryVideoEvent, StoryVideoProject } from "../../api_storyvideo";
import { StoryVideoAudioPanel } from "./StoryVideoAudioPanel";
import { StoryVideoDraftEditor } from "./StoryVideoDraftEditor";
import { StoryVideoEvents } from "./StoryVideoEvents";
import { StoryVideoShotsPanel } from "./StoryVideoShotsPanel";
import { StoryVideoVideoPanel } from "./StoryVideoVideoPanel";

const STEP_TITLES = ["草稿确认", "分镜图片", "解说音频", "最终成片", "过程日志"] as const;

export function StoryVideoWorkspace(props: { project: StoryVideoProject | null; events: StoryVideoEvent[]; busy: boolean; error: string }) {
  const [active, setActive] = useState(0);
  useEffect(() => setActive(0), [props.project?.id]);
  if (!props.project) return <section className="card"><div className="placeholder"><div className="placeholder__title">先创建或选择一个故事项目</div><div className="placeholder__sub">创建后会在这里按步骤展示草稿、分镜、音频、视频和日志。</div></div></section>;
  const project = props.project;
  const steps = [
    <StoryVideoDraftEditor key="draft" project={project} busy={props.busy} />,
    <StoryVideoShotsPanel key="shots" project={project} busy={props.busy} />,
    <StoryVideoAudioPanel key="audio" project={project} busy={props.busy} error={props.error} />,
    <StoryVideoVideoPanel key="video" project={project} busy={props.busy} error={props.error} />,
    <StoryVideoEvents key="events" events={props.events} />
  ];
  function goTo(index: number) {
    setActive(Math.max(0, Math.min(index, steps.length - 1)));
  }
  function onWheel(event: React.WheelEvent<HTMLDivElement>) {
    if (Math.abs(event.deltaX) < 24 && Math.abs(event.deltaY) < 48) return;
    if (Math.abs(event.deltaX) >= Math.abs(event.deltaY)) {
      event.preventDefault();
      goTo(active + (event.deltaX > 0 ? 1 : -1));
    }
  }
  function onKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === "ArrowRight") goTo(active + 1);
    if (event.key === "ArrowLeft") goTo(active - 1);
  }
  return (
    <div className="storyvideoWorkspace">
      <section className="card storyvideoSummary">
        <div className="storyvideoSummary__head">
          <div>
            <h2 className="card__title">{project.title || "未命名项目"}</h2>
            <div className="storyvideoSummary__meta">{(project.keywords || []).join(" / ") || "暂无关键词"}</div>
          </div>
          <span className="badge">{project.status}</span>
        </div>
        <div className="storyvideoSummary__grid">
          <div className="panel"><div className="k">摘要</div><div>{project.summary || "暂无摘要"}</div></div>
          <div className="panel"><div className="k">模型链路</div><div>{project.planner_provider} / {project.image_provider} / {project.audio_provider}</div></div>
          <div className="panel"><div className="k">时长与画幅</div><div>{project.duration_seconds}s · {project.aspect_ratio || "16:9"}</div></div>
        </div>
      </section>
      <section className="card storyvideoRail">
        <div className="storyvideoRail__head">
          <div className="storyvideoRail__steps" role="tablist" aria-label="故事工坊步骤">
            {STEP_TITLES.map((title, index) => <button key={title} role="tab" aria-selected={index === active} className={index === active ? "storyvideoRail__step storyvideoRail__step--active" : "storyvideoRail__step"} onClick={() => goTo(index)}>{index + 1}. {title}</button>)}
          </div>
          <div className="storyvideoRail__actions">
            <button className="btn btn--ghost" disabled={active === 0} onClick={() => goTo(active - 1)}>上一项</button>
            <button className="btn btn--ghost" disabled={active === steps.length - 1} onClick={() => goTo(active + 1)}>下一项</button>
          </div>
        </div>
        <div className="storyvideoRail__hint">使用顶部步骤、左右按钮，或触控板左右滑动来切换。</div>
        <div className="storyvideoRail__viewport" tabIndex={0} onWheel={onWheel} onKeyDown={onKeyDown}>
          <div className="storyvideoRail__track" style={{ transform: `translateX(-${active * 100}%)` }}>
            {steps.map((step, index) => <div key={STEP_TITLES[index]} className="storyvideoRail__slide">{step}</div>)}
          </div>
        </div>
      </section>
    </div>
  );
}
