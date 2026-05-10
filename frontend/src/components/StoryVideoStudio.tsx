import React, { useEffect } from "react";
import { useStoryVideo } from "../state/storyvideo";
import { StoryVideoForm } from "./storyvideo/StoryVideoForm";
import { StoryVideoProjects } from "./storyvideo/StoryVideoProjects";
import { StoryVideoWorkspace } from "./storyvideo/StoryVideoWorkspace";

export function StoryVideoStudio(props: { mode: "create" | "records"; onModeChange: (mode: "create" | "records") => void }) {
  const { project, projects, events, imageMeta, audioMeta, textMeta, busy, error, selectProject } = useStoryVideo();
  const recordsMode = props.mode === "records";
  useEffect(() => {
    if (!recordsMode || project?.id || !projects[0]?.id || busy) return;
    void selectProject(projects[0].id);
  }, [busy, project?.id, projects, recordsMode, selectProject]);
  return (
    <div className="storyvideo">
      <div className="chips">
        <button className={recordsMode ? "chip chip--ghost" : "chip"} onClick={() => props.onModeChange("create")} disabled={busy}><span className="chip__text">新建项目</span></button>
        <button className={recordsMode ? "chip" : "chip chip--ghost"} onClick={() => props.onModeChange("records")} disabled={busy}><span className="chip__text">项目记录</span></button>
      </div>
      {recordsMode ? (
        <div className="storyvideo__records">
          <StoryVideoProjects projects={projects} currentId={project?.id || ""} busy={busy} onSelect={selectProject} />
          <StoryVideoWorkspace project={project} events={events} busy={busy} error={error} />
        </div>
      ) : (
        <div className="storyvideo__create">
          <StoryVideoForm imageMeta={imageMeta} audioMeta={audioMeta} textMeta={textMeta} busy={busy} error={error} />
          <StoryVideoWorkspace project={project} events={events} busy={busy} error={error} />
        </div>
      )}
    </div>
  );
}
