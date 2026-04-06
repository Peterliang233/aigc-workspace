import React, { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";
import { animationApi, type AnimationJobCreateRequest, type AnimationJobGetResponse } from "../api_animation";

type AnimationTask = AnimationJobGetResponse & { created_at: number };
type State = { jobs: AnimationTask[] };
type Ctx = {
  state: State;
  createAnimationJob: (req: AnimationJobCreateRequest) => Promise<AnimationJobCreateResponse>;
  removeAnimationJob: (jobID: string) => void;
};
type AnimationJobCreateResponse = { job_id: string; status: string; provider: string; model?: string; duration_seconds: number };

const LS_KEY = "aigc_animations_v1";
const MAX_KEEP = 20;
const AnimationContext = createContext<Ctx | null>(null);

function readState(): State {
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (!raw) return { jobs: [] };
    const obj = JSON.parse(raw);
    return { jobs: Array.isArray(obj?.jobs) ? obj.jobs : [] };
  } catch {
    return { jobs: [] };
  }
}

export function AnimationProvider(props: { children: React.ReactNode }) {
  const [state, setState] = useState<State>(() => readState());
  const ref = useRef(state);
  useEffect(() => {
    ref.current = state;
    localStorage.setItem(LS_KEY, JSON.stringify(state));
  }, [state]);

  useEffect(() => {
    let stop = false;
    let timer: number | null = null;
    async function tick() {
      const pending = ref.current.jobs.filter((j) => j.status !== "succeeded" && j.status !== "failed");
      for (const job of pending) {
        try {
          const res = await animationApi.getJob(job.job_id);
          setState((prev) => ({ ...prev, jobs: prev.jobs.map((x) => (x.job_id === job.job_id ? { ...x, ...res } : x)) }));
        } catch {
          // ignore transient poll errors
        }
      }
      if (!stop) timer = window.setTimeout(tick, 2500);
    }
    timer = window.setTimeout(tick, 1000);
    return () => {
      stop = true;
      if (timer) window.clearTimeout(timer);
    };
  }, []);

  async function createAnimationJob(req: AnimationJobCreateRequest) {
    const res = await animationApi.createJob(req);
    const task: AnimationTask = { ...res, created_at: Date.now(), segment_count: 0, completed_segments: 0, current_segment: 0, segments: [] };
    setState((prev) => ({ ...prev, jobs: [task, ...prev.jobs].slice(0, MAX_KEEP) }));
    return res;
  }

  function removeAnimationJob(jobID: string) {
    setState((prev) => ({ ...prev, jobs: prev.jobs.filter((j) => j.job_id !== jobID) }));
  }

  const value = useMemo(() => ({ state, createAnimationJob, removeAnimationJob }), [state]);
  return <AnimationContext.Provider value={value}>{props.children}</AnimationContext.Provider>;
}

export function useAnimationJobs() {
  const ctx = useContext(AnimationContext);
  if (!ctx) throw new Error("useAnimationJobs must be used within <AnimationProvider />");
  return ctx;
}
