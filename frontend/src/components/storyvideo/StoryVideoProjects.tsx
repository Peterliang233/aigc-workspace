import React from "react";
import type { StoryVideoProject } from "../../api_storyvideo";

export function StoryVideoProjects(props: { projects: StoryVideoProject[]; currentId: string; busy: boolean; onSelect: (id: string) => Promise<void> }) {
  return (
    <section className="card storyvideoProjects">
      <div className="card__head">
        <h2 className="card__title">项目记录</h2>
      </div>
      <div className="list">
        {props.projects.map((item) => (
          <button key={item.id} className={item.id === props.currentId ? "hrow hrow--active" : "hrow"} disabled={props.busy} onClick={() => void props.onSelect(item.id)}>
            <div className="hrow__top">
              <strong>{item.title || item.id}</strong>
              <span className="pill">{item.status}</span>
            </div>
            <div className="hrow__bot">
              <div className="hrow__prompt">{item.summary || item.keywords?.join(" / ") || "暂无摘要"}</div>
              <div className="k">{item.planner_provider} · {item.image_provider} · {item.audio_provider}</div>
            </div>
          </button>
        ))}
        {props.projects.length === 0 ? <div className="panel">还没有故事视频项目。</div> : null}
      </div>
    </section>
  );
}
