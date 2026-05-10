export type StoryVideoDraftRequest = {
  keywords: string[];
  theme?: string;
  audience?: string;
  tone?: string;
  extra?: string;
  duration_seconds: number;
  aspect_ratio?: string;
  planner_provider?: string;
  planner_model?: string;
  image_provider?: string;
  image_model?: string;
  audio_provider?: string;
  audio_model?: string;
  audio_voice?: string;
};

export type StoryVideoShot = {
  id: string;
  index: number;
  title: string;
  story_beat: string;
  narration_line: string;
  image_prompt: string;
  image_url?: string;
  audio_url?: string;
  duration_ms: number;
  status: string;
  attempt_count: number;
  error?: string;
};

export type StoryVideoProject = {
  id: string;
  status: string;
  keywords: string[];
  theme?: string;
  audience?: string;
  tone?: string;
  extra?: string;
  duration_seconds: number;
  aspect_ratio?: string;
  title?: string;
  summary?: string;
  script_text?: string;
  narration_text?: string;
  planner_provider?: string;
  planner_model?: string;
  image_provider?: string;
  image_model?: string;
  audio_provider?: string;
  audio_model?: string;
  audio_voice?: string;
  audio_url?: string;
  video_url?: string;
  error?: string;
  created_at?: string;
  updated_at?: string;
  shots?: StoryVideoShot[];
};

export type StoryVideoEvent = {
  id: number;
  stage: string;
  type: string;
  message: string;
  payload?: string;
  created_at: string;
};

export type StoryVideoDraftUpdateRequest = Omit<StoryVideoProject, "id" | "status" | "audio_url" | "video_url" | "error" | "shots"> & {
  shots: Array<{
    id?: string;
    title: string;
    story_beat: string;
    narration_line: string;
    image_prompt: string;
    duration_ms: number;
  }>;
};

async function httpJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {})
    }
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = typeof data?.error === "string" ? data.error : `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return data as T;
}

export const storyVideoApi = {
  createDraft: (req: StoryVideoDraftRequest) =>
    httpJSON<StoryVideoProject>("/api/story-videos/projects/draft", { method: "POST", body: JSON.stringify(req) }),
  updateDraft: (projectId: string, req: StoryVideoDraftUpdateRequest) =>
    httpJSON<StoryVideoProject>(`/api/story-videos/projects/${encodeURIComponent(projectId)}/draft`, { method: "PUT", body: JSON.stringify(req) }),
  confirm: (projectId: string) =>
    httpJSON<StoryVideoProject>(`/api/story-videos/projects/${encodeURIComponent(projectId)}/confirm`, { method: "POST" }),
  list: () => httpJSON<{ items: StoryVideoProject[] }>("/api/story-videos/projects", { method: "GET" }),
  get: (projectId: string) =>
    httpJSON<StoryVideoProject>(`/api/story-videos/projects/${encodeURIComponent(projectId)}`, { method: "GET" }),
  events: (projectId: string) =>
    httpJSON<{ items: StoryVideoEvent[] }>(`/api/story-videos/projects/${encodeURIComponent(projectId)}/events`, { method: "GET" }),
  regenerateAudio: (projectId: string, payload?: { narration_text?: string; audio_provider?: string; audio_model?: string; audio_voice?: string }) =>
    httpJSON<StoryVideoProject>(`/api/story-videos/projects/${encodeURIComponent(projectId)}/regenerate-audio`, { method: "POST", body: JSON.stringify(payload || {}) }),
  regenerateShot: (projectId: string, shotId: string, payload?: { image_prompt?: string; image_provider?: string; image_model?: string }) =>
    httpJSON<StoryVideoProject>(`/api/story-videos/projects/${encodeURIComponent(projectId)}/shots/${encodeURIComponent(shotId)}/regenerate-image`, { method: "POST", body: JSON.stringify(payload || {}) }),
  compose: (projectId: string) =>
    httpJSON<StoryVideoProject>(`/api/story-videos/projects/${encodeURIComponent(projectId)}/compose`, { method: "POST" })
};
