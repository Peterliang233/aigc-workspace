import React, { useMemo } from "react";
import { useGeneration } from "../state/generation";
import { VideoJobForm } from "./video/VideoJobForm";
import { VideoResults } from "./video/VideoResults";

export function VideoStudio() {
  const { state } = useGeneration();
  const jobs = state.videos;
  const latest = useMemo(() => jobs[0] || null, [jobs]);

  return (
    <div className="workspace">
      <VideoJobForm latestStatus={latest?.status} />
      <VideoResults jobs={jobs} />
    </div>
  );
}
