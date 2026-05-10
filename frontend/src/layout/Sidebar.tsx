import React from "react";
import { Icon } from "./Icon";
import { StoryVideoMark } from "./StoryVideoMark";
import type { Tab } from "./tabs";

export function Sidebar(props: {
  tab: Tab;
  sideCollapsed: boolean;
  sideOpen: boolean;
  onSelect: (t: Tab) => void;
  onToggleCollapsed: () => void;
  onCloseDrawer: () => void;
}) {
  const { tab, sideCollapsed, sideOpen, onSelect, onToggleCollapsed, onCloseDrawer } = props;
  const storyActive = tab === "animation-create" || tab === "animation-records";
  return (
    <aside
      className={[
        "side",
        sideCollapsed ? "side--collapsed" : "",
        sideOpen ? "side--open" : ""
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <div className="side__top">
        <button
          className="sideToggle"
          onClick={onToggleCollapsed}
          title={sideCollapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          <Icon name={sideCollapsed ? "expand" : "collapse"} />
        </button>

        <button className="sideToggle sideToggle--mobile" onClick={onCloseDrawer} title="Close menu">
          <Icon name="collapse" />
        </button>
      </div>

      <div className="side__brand">
        <div className="brand__title">
          <span className="brand__titleFull">AIGC Studio</span>
          <span className="brand__titleMini">AIGC</span>
        </div>
      </div>

      <nav className="side__nav">
        <button
          className={tab === "text" ? "navitem navitem--active" : "navitem"}
          onClick={() => onSelect("text")}
          title="文本生成"
        >
          <span className="navitem__icon">
            <Icon name="text" />
          </span>
          <span className="navitem__label">文本生成</span>
        </button>
        <button
          className={tab === "image" ? "navitem navitem--active" : "navitem"}
          onClick={() => onSelect("image")}
          title="图片生成"
        >
          <span className="navitem__icon">
            <Icon name="image" />
          </span>
          <span className="navitem__label">图片生成</span>
        </button>
        <button
          className={tab === "video" ? "navitem navitem--active" : "navitem"}
          onClick={() => onSelect("video")}
          title="视频生成"
        >
          <span className="navitem__icon">
            <Icon name="video" />
          </span>
          <span className="navitem__label">视频生成</span>
        </button>
        <button
          className={tab === "audio" ? "navitem navitem--active" : "navitem"}
          onClick={() => onSelect("audio")}
          title="音频生成"
        >
          <span className="navitem__icon">
            <Icon name="audio" />
          </span>
          <span className="navitem__label">音频生成</span>
        </button>
        <div className={storyActive ? "navgroup navgroup--active" : "navgroup"}>
          <button className={storyActive ? "navitem navitem--active" : "navitem"} onClick={() => onSelect("animation-create")} title="故事视频工坊">
            <span className="navitem__icon">
              <StoryVideoMark />
            </span>
            <span className="navitem__label">故事视频工坊</span>
          </button>
          <div className="navgroup__sub">
            <button className={tab === "animation-create" ? "navsub navsub--active" : "navsub"} onClick={() => onSelect("animation-create")}>
              <span className="navsub__dot" />
              <span className="navsub__label">新建项目</span>
            </button>
            <button className={tab === "animation-records" ? "navsub navsub--active" : "navsub"} onClick={() => onSelect("animation-records")}>
              <span className="navsub__dot" />
              <span className="navsub__label">项目记录</span>
            </button>
          </div>
        </div>
        <button
          className={tab === "toolbox" ? "navitem navitem--active" : "navitem"}
          onClick={() => onSelect("toolbox")}
          title="工具箱"
        >
          <span className="navitem__icon">
            <Icon name="toolbox" />
          </span>
          <span className="navitem__label">工具箱</span>
        </button>
        <button
          className={tab === "history" ? "navitem navitem--active" : "navitem"}
          onClick={() => onSelect("history")}
          title="历史"
        >
          <span className="navitem__icon">
            <Icon name="history" />
          </span>
          <span className="navitem__label">历史</span>
        </button>
      </nav>
    </aside>
  );
}
