import React, { useMemo, useState } from "react";
import { ImageStudio } from "./components/ImageStudio";
import { VideoStudio } from "./components/VideoStudio";

type Tab = "image" | "video";

function Icon({ name }: { name: "image" | "video" | "menu" | "collapse" | "expand" }) {
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

  const subtitle = useMemo(() => {
    if (tab === "image") return "生成图片: Prompt -> Image";
    return "生成视频: Prompt -> Async Job -> Video";
  }, [tab]);

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
            <div className="brand__sub">{subtitle}</div>
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
          </nav>

          <div className="side__foot">
            <div className="side__footrow">
              <span className="muter">Backend</span>
              <a href="/healthz" target="_blank" rel="noreferrer" title="/healthz">
                /healthz
              </a>
            </div>
            <div className="side__footrow">
              <span className="muter">Provider</span>
              <span className="muted2" title="backend/.env">
                backend/.env
              </span>
            </div>
          </div>
        </aside>

        <main className="main">
          <div className="mobilebar">
            <button className="mobilebar__btn" onClick={() => setSideOpen(true)} title="Open menu">
              <Icon name="menu" />
            </button>
            <div className="mobilebar__title">{tab === "image" ? "图片生成" : "视频生成"}</div>
          </div>

          {tab === "image" ? <ImageStudio /> : <VideoStudio />}
        </main>
      </div>
    </div>
  );
}
