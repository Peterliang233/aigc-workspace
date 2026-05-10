import React, { createContext, useContext, useEffect, useMemo, useState } from "react";
import { api, type ProviderMetaResponse } from "../api";
import { storyVideoApi, type StoryVideoDraftRequest, type StoryVideoDraftUpdateRequest, type StoryVideoEvent, type StoryVideoProject } from "../api_storyvideo";

type Ctx = {
  projects: StoryVideoProject[];
  project: StoryVideoProject | null;
  events: StoryVideoEvent[];
  imageMeta: ProviderMetaResponse | null;
  audioMeta: ProviderMetaResponse | null;
  busy: boolean;
  error: string;
  selectProject: (id: string) => Promise<void>;
  createDraft: (req: StoryVideoDraftRequest) => Promise<void>;
  saveDraft: (req: StoryVideoDraftUpdateRequest) => Promise<void>;
  confirmProject: () => Promise<void>;
  regenerateAudio: (payload?: { narration_text?: string; audio_provider?: string; audio_model?: string }) => Promise<void>;
  regenerateShot: (shotId: string, payload?: { image_prompt?: string; image_provider?: string; image_model?: string }) => Promise<void>;
  composeProject: () => Promise<void>;
};

const StoryVideoContext = createContext<Ctx | null>(null);

export function StoryVideoProvider(props: { children: React.ReactNode }) {
  const [projects, setProjects] = useState<StoryVideoProject[]>([]);
  const [project, setProject] = useState<StoryVideoProject | null>(null);
  const [events, setEvents] = useState<StoryVideoEvent[]>([]);
  const [imageMeta, setImageMeta] = useState<ProviderMetaResponse | null>(null);
  const [audioMeta, setAudioMeta] = useState<ProviderMetaResponse | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function refreshProjects() {
    const res = await storyVideoApi.list();
    setProjects(res.items || []);
  }

  async function loadProject(id: string) {
    const [nextProject, nextEvents] = await Promise.all([storyVideoApi.get(id), storyVideoApi.events(id)]);
    setProject(nextProject);
    setEvents(nextEvents.items || []);
  }

  useEffect(() => {
    void (async () => {
      try {
        const [images, audios] = await Promise.all([api.getImageMeta(), api.getAudioMeta()]);
        setImageMeta(images);
        setAudioMeta(audios);
        await refreshProjects();
      } catch (e: any) {
        setError(e?.message || String(e));
      }
    })();
  }, []);

  useEffect(() => {
    if (!project?.id) return;
    const timer = window.setInterval(() => void loadProject(project.id).catch(() => {}), 2500);
    return () => window.clearInterval(timer);
  }, [project?.id]);

  async function withBusy(fn: () => Promise<void>) {
    setBusy(true);
    setError("");
    try {
      await fn();
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  const value = useMemo<Ctx>(() => ({
    projects,
    project,
    events,
    imageMeta,
    audioMeta,
    busy,
    error,
    selectProject: async (id: string) => withBusy(async () => loadProject(id)),
    createDraft: async (req) => withBusy(async () => {
      const next = await storyVideoApi.createDraft(req);
      setProject(next);
      await refreshProjects();
      const ev = await storyVideoApi.events(next.id);
      setEvents(ev.items || []);
    }),
    saveDraft: async (req) => withBusy(async () => {
      if (!project) return;
      const next = await storyVideoApi.updateDraft(project.id, req);
      setProject(next);
      await refreshProjects();
    }),
    confirmProject: async () => withBusy(async () => {
      if (!project) return;
      setProject(await storyVideoApi.confirm(project.id));
      await refreshProjects();
    }),
    regenerateAudio: async (payload) => withBusy(async () => {
      if (!project) return;
      setProject(await storyVideoApi.regenerateAudio(project.id, payload));
    }),
    regenerateShot: async (shotId, payload) => withBusy(async () => {
      if (!project) return;
      setProject(await storyVideoApi.regenerateShot(project.id, shotId, payload));
    }),
    composeProject: async () => withBusy(async () => {
      if (!project) return;
      setProject(await storyVideoApi.compose(project.id));
    })
  }), [projects, project, events, imageMeta, audioMeta, busy, error]);

  return <StoryVideoContext.Provider value={value}>{props.children}</StoryVideoContext.Provider>;
}

export function useStoryVideo() {
  const ctx = useContext(StoryVideoContext);
  if (!ctx) throw new Error("useStoryVideo must be used within <StoryVideoProvider />");
  return ctx;
}
