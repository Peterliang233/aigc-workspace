import React, { useState } from "react";
import { ImageStudio } from "./components/ImageStudio";
import { VideoStudio } from "./components/VideoStudio";
import { ConfigStudio } from "./components/ConfigStudio";
import { HistoryStudio } from "./components/HistoryStudio";
import { ToolboxStudio } from "./components/ToolboxStudio";

import { Sidebar } from "./layout/Sidebar";
import { MobileBar } from "./layout/MobileBar";
import type { Tab } from "./layout/tabs";

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

        <Sidebar
          tab={tab}
          sideCollapsed={sideCollapsed}
          sideOpen={sideOpen}
          onSelect={onSelect}
          onToggleCollapsed={() => setSideCollapsed((v) => !v)}
          onCloseDrawer={() => setSideOpen(false)}
        />

        <main className="main">
          <MobileBar tab={tab} onOpenMenu={() => setSideOpen(true)} />

          {tab === "image" ? (
            <ImageStudio />
          ) : tab === "video" ? (
            <VideoStudio />
          ) : tab === "config" ? (
            <ConfigStudio />
          ) : tab === "toolbox" ? (
            <ToolboxStudio />
          ) : (
            <HistoryStudio />
          )}
        </main>
      </div>
    </div>
  );
}
