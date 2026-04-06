import React, { useMemo } from "react";
import { useAnimationJobs } from "../state/animation";
import { AnimationJobForm } from "./animation/AnimationJobForm";
import { AnimationResults } from "./animation/AnimationResults";

export function AnimationStudio() {
  const { state, removeAnimationJob } = useAnimationJobs();
  const jobs = state.jobs;
  const latest = useMemo(() => jobs[0] || null, [jobs]);
  return (
    <div className="workspace">
      <AnimationJobForm latestStatus={latest?.status} />
      <AnimationResults
        jobs={jobs}
        onDeleteJob={(jobID) => {
          const row = jobs.find((x) => x.job_id === jobID);
          if (!row || row.status === "succeeded") return;
          if (!window.confirm(`确定删除动画任务 ${jobID} 吗？`)) return;
          removeAnimationJob(jobID);
        }}
      />
    </div>
  );
}
