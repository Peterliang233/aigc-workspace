import React, { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { AudioGenerateRequest, ImageGenerateRequest, VideoJobCreateRequest, VideoJobGetResponse } from "../api";
import { api } from "../api";
import { recoverRunningImages } from "./imageRecovery";
import { makeId, persistState, safeParseState } from "./generationState";

type ImageTask = {
  id: string;
  status: "running" | "succeeded" | "failed";
  provider?: string;
  model?: string;
  prompt: string;
  size?: string;
  image_url?: string;
  error?: string;
  created_at: number;
};

type VideoTask = VideoJobGetResponse & { created_at: number };
type AudioTask = {
  id: string;
  status: "running" | "succeeded" | "failed";
  provider?: string;
  model?: string;
  voice?: string;
  input: string;
  audio_url?: string;
  error?: string;
  created_at: number;
};

type GenState = {
  images: ImageTask[];
  videos: VideoTask[];
  audios: AudioTask[];
};

type Ctx = {
  state: GenState;
  startImage: (req: ImageGenerateRequest) => string;
  startAudio: (req: AudioGenerateRequest) => string;
  createVideoJob: (req: VideoJobCreateRequest) => Promise<{ job_id: string; status: string; provider: string }>;
  removeVideoJob: (jobID: string) => void;
};

const MAX_KEEP = 20;

const GenerationContext = createContext<Ctx | null>(null);

export function GenerationProvider(props: { children: React.ReactNode }) {
  const [state, setState] = useState<GenState>(() => safeParseState());
  const stateRef = useRef(state);
  useEffect(() => {
    stateRef.current = state;
  }, [state]);

  useEffect(() => {
    persistState(state);
  }, [state]);

  // Poll video jobs globally so tab switches do not stop polling.
  useEffect(() => {
    let stopped = false;
    let t: number | null = null;

    async function tick() {
      if (stopped) return;
      const inflight = stateRef.current.videos.filter((v) => v.status !== "succeeded" && v.status !== "failed");
      for (const v of inflight) {
        try {
          const res = await api.getVideoJob(v.job_id);
          setState((prev) => ({
            ...prev,
            videos: prev.videos.map((x) => (x.job_id === res.job_id ? { ...x, ...res } : x))
          }));
        } catch {
          // ignore transient polling errors
        }
      }
      t = window.setTimeout(tick, 1800);
    }

    t = window.setTimeout(tick, 900);
    return () => {
      stopped = true;
      if (t) window.clearTimeout(t);
    };
  }, []);

  useEffect(() => {
    let stopped = false;
    let t: number | null = null;

    async function tick() {
      const snapshot = stateRef.current.images;
      const next = await recoverRunningImages(snapshot);
      if (!stopped && next !== snapshot) {
        setState((prev) => (prev.images === snapshot ? { ...prev, images: next } : prev));
      }
      if (stopped) return;
      t = window.setTimeout(tick, 4000);
    }

    t = window.setTimeout(tick, 1200);
    return () => {
      stopped = true;
      if (t) window.clearTimeout(t);
    };
  }, []);

  function startImage(req: ImageGenerateRequest) {
    const id = makeId();
    const task: ImageTask = {
      id,
      status: "running",
      provider: req.provider,
      model: req.model,
      prompt: req.prompt,
      size: req.size,
      created_at: Date.now()
    };

    setState((prev) => ({ ...prev, images: [task, ...prev.images].slice(0, MAX_KEEP) }));

    (async () => {
      try {
        const res = await api.generateImage(req);
        const next = (res.image_urls || [])[0] || "";
        setState((prev) => ({
          ...prev,
          images: prev.images.map((x) =>
            x.id === id
              ? {
                  ...x,
                  status: next ? "succeeded" : "failed",
                  image_url: next || undefined,
                  error: next ? undefined : "empty image url from provider"
                }
              : x
          )
        }));
      } catch (e: any) {
        setState((prev) => ({
          ...prev,
          images: prev.images.map((x) => (x.id === id ? { ...x, status: "failed", error: e?.message || String(e) } : x))
        }));
      }
    })();

    return id;
  }

  function startAudio(req: AudioGenerateRequest) {
    const id = makeId();
    const task: AudioTask = { id, status: "running", provider: req.provider, model: req.model, voice: req.voice, input: req.input, created_at: Date.now() };
    setState((prev) => ({ ...prev, audios: [task, ...prev.audios].slice(0, MAX_KEEP) }));
    (async () => {
      try {
        const res = await api.generateAudio(req);
        setState((prev) => ({
          ...prev,
          audios: prev.audios.map((x) => x.id === id ? { ...x, status: res.audio_url ? "succeeded" : "failed", audio_url: res.audio_url || undefined, model: res.model || x.model, error: res.audio_url ? undefined : "empty audio url from provider" } : x)
        }));
      } catch (e: any) {
        setState((prev) => ({
          ...prev,
          audios: prev.audios.map((x) => x.id === id ? { ...x, status: "failed", error: e?.message || String(e) } : x)
        }));
      }
    })();
    return id;
  }

  async function createVideoJob(req: VideoJobCreateRequest) {
    const res = await api.createVideoJob(req);
    const row: VideoTask = { job_id: res.job_id, status: res.status, provider: res.provider, created_at: Date.now() };
    setState((prev) => ({ ...prev, videos: [row, ...prev.videos].slice(0, MAX_KEEP) }));
    return res;
  }

  function removeVideoJob(jobID: string) {
    jobID = String(jobID || "").trim();
    if (!jobID) return;
    setState((prev) => ({ ...prev, videos: prev.videos.filter((x) => x.job_id !== jobID) }));
  }

  const value = useMemo<Ctx>(() => ({ state, startImage, startAudio, createVideoJob, removeVideoJob }), [state]);
  return <GenerationContext.Provider value={value}>{props.children}</GenerationContext.Provider>;
}

export function useGeneration() {
  const ctx = useContext(GenerationContext);
  if (!ctx) throw new Error("useGeneration must be used within <GenerationProvider />");
  return ctx;
}
