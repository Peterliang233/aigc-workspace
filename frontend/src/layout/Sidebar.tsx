import React from "react";
import { Icon } from "./Icon";
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
