import React, { useState } from "react";
import { AudioStudio } from "./components/AudioStudio";
import { ImageStudio } from "./components/ImageStudio";
import { TextStudio } from "./components/TextStudio";
import { StoryVideoStudio } from "./components/StoryVideoStudio";
import { VideoStudio } from "./components/VideoStudio";
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

          {tab === "text" ? (
            <TextStudio />
          ) : tab === "image" ? (
            <ImageStudio />
          ) : tab === "video" ? (
            <VideoStudio />
          ) : tab === "audio-tts" ? (
            <AudioStudio category="文本转语音" title="文本转语音" />
          ) : tab === "audio-speech" ? (
            <AudioStudio category="语音合成" title="语音合成" />
          ) : tab === "animation-create" || tab === "animation-records" ? (
            <StoryVideoStudio
              mode={tab === "animation-records" ? "records" : "create"}
              onModeChange={(mode) => onSelect(mode === "records" ? "animation-records" : "animation-create")}
            />
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
