import React, { useMemo } from "react";
import { useGeneration } from "../state/generation";
import { VideoJobForm } from "./video/VideoJobForm";
import { VideoResults } from "./video/VideoResults";

export function VideoStudio() {
  const { state, removeVideoJob } = useGeneration();
  const jobs = state.videos;
  const latest = useMemo(() => jobs[0] || null, [jobs]);

  return (
    <div className="workspace">
      <VideoJobForm latestStatus={latest?.status} />
      <VideoResults
        jobs={jobs}
        onDeleteJob={(jobID) => {
          const t = jobs.find((x) => x.job_id === jobID);
          if (!t || t.status === "succeeded") return;
          if (!window.confirm(`确定删除任务 ${jobID} 吗？`)) return;
          removeVideoJob(jobID);
        }}
      />
    </div>
  );
}
