import React, { useState } from "react";
import { ImageStudio } from "./components/ImageStudio";
import { VideoStudio } from "./components/VideoStudio";
import { ConfigStudio } from "./components/ConfigStudio";

type Tab = "image" | "video" | "config";

function Icon({ name }: { name: "image" | "video" | "settings" | "menu" | "collapse" | "expand" }) {
  if (name === "menu") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M3 5h14M3 10h14M3 15h14"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
        />
      </svg>
    );
  }
  if (name === "collapse") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M12 4l-6 6 6 6"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }
  if (name === "expand") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M8 4l6 6-6 6"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }
  if (name === "settings") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M8.2 3.4h3.6l.4 2.1a6.3 6.3 0 0 1 1.4.8l2-.8 1.8 3.1-1.6 1.4c.1.3.1.7.1 1.1s0 .8-.1 1.1l1.6 1.4-1.8 3.1-2-.8c-.4.3-.9.6-1.4.8l-.4 2.1H8.2l-.4-2.1a6.3 6.3 0 0 1-1.4-.8l-2 .8-1.8-3.1 1.6-1.4A4.8 4.8 0 0 1 4.8 10c0-.4 0-.8.1-1.1L3.3 7.5 5.1 4.4l2 .8c.4-.3.9-.6 1.4-.8l.4-2.1Z"
          stroke="currentColor"
          strokeWidth="1.2"
          strokeLinejoin="round"
        />
        <path
          d="M10 12.6a2.6 2.6 0 1 0 0-5.2 2.6 2.6 0 0 0 0 5.2Z"
          stroke="currentColor"
          strokeWidth="1.2"
        />
      </svg>
    );
  }
  if (name === "image") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M4.5 5.8c0-.72.58-1.3 1.3-1.3h8.4c.72 0 1.3.58 1.3 1.3v8.4c0 .72-.58 1.3-1.3 1.3H5.8c-.72 0-1.3-.58-1.3-1.3V5.8Z"
          stroke="currentColor"
          strokeWidth="1.4"
        />
        <path
          d="M6.3 12.4l2.2-2.2 2.2 2.2 1.3-1.3 2.0 2.0"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M7.6 8.2h.01"
          stroke="currentColor"
          strokeWidth="2.6"
          strokeLinecap="round"
        />
      </svg>
    );
  }
  // video
  return (
    <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <path
        d="M4.8 6.2c0-.66.54-1.2 1.2-1.2h6.2c.66 0 1.2.54 1.2 1.2v7.6c0 .66-.54 1.2-1.2 1.2H6c-.66 0-1.2-.54-1.2-1.2V6.2Z"
        stroke="currentColor"
        strokeWidth="1.4"
      />
      <path
        d="M13.4 8l2.8-1.6v7.2L13.4 12V8Z"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function App() {
  const [tab, setTab] = useState<Tab>("image");
  const [sideCollapsed, setSideCollapsed] = useState(false);
  const [sideOpen, setSideOpen] = useState(false); // mobile drawer

  function onSelect(next: Tab) {
    setTab(next);
    setSideOpen(false);
  }

  return (
    <div className="page">
      <div className={sideOpen ? "shell shell--drawer-open" : "shell"}>
        <div
          className={sideOpen ? "backdrop backdrop--show" : "backdrop"}
          onClick={() => setSideOpen(false)}
        />

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
              onClick={() => setSideCollapsed((v) => !v)}
              title={sideCollapsed ? "Expand sidebar" : "Collapse sidebar"}
            >
              <Icon name={sideCollapsed ? "expand" : "collapse"} />
            </button>

            <button className="sideToggle sideToggle--mobile" onClick={() => setSideOpen(false)} title="Close menu">
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
              className={tab === "config" ? "navitem navitem--active" : "navitem"}
              onClick={() => onSelect("config")}
              title="配置"
            >
              <span className="navitem__icon">
                <Icon name="settings" />
              </span>
              <span className="navitem__label">配置</span>
            </button>
          </nav>
        </aside>

        <main className="main">
          <div className="mobilebar">
            <button className="mobilebar__btn" onClick={() => setSideOpen(true)} title="Open menu">
              <Icon name="menu" />
            </button>
            <div className="mobilebar__title">
              {tab === "image" ? "图片生成" : tab === "video" ? "视频生成" : "配置"}
            </div>
          </div>

          {tab === "image" ? <ImageStudio /> : tab === "video" ? <VideoStudio /> : <ConfigStudio />}
        </main>
      </div>
    </div>
  );
}
