export type ImageGenerateRequest = {
  provider?: string;
  model?: string;
  prompt: string;
  size?: string;
  n?: number;
};

export type ImageGenerateResponse = {
  image_urls: string[];
  provider: string;
  model?: string;
};

export type VideoJobCreateRequest = {
  prompt: string;
  duration_seconds?: number;
  aspect_ratio?: string;
};

export type VideoJobCreateResponse = {
  job_id: string;
  status: string;
  provider: string;
};

export type VideoJobGetResponse = {
  job_id: string;
  status: string;
  video_url?: string;
  error?: string;
  provider: string;
};

export type SettingsGetResponse = {
  image_providers: Record<
    string,
    {
      label: string;
      base_url?: string;
      api_key_set: boolean;
      default_model?: string;
      models?: string[];
    }
  >;
};

export type ProviderSettingsPatch = {
  base_url?: string;
  api_key?: string;
  default_model?: string;
  models?: string[];
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

export const api = {
  getImageMeta: () =>
    httpJSON<{
      default_provider: string;
      providers: { id: string; label: string; configured: boolean; models: string[] }[];
    }>("/api/meta/images", { method: "GET" }),

  getSettings: () => httpJSON<SettingsGetResponse>("/api/settings", { method: "GET" }),

  updateSettings: (patch: { image_providers: Record<string, ProviderSettingsPatch> }) =>
    httpJSON<{ ok: boolean }>("/api/settings", {
      method: "PUT",
      body: JSON.stringify(patch)
    }),

  addImageModel: (providerId: string, model: string) =>
    httpJSON<{ ok: boolean }>(`/api/settings/image-providers/${encodeURIComponent(providerId)}/models`, {
      method: "POST",
      body: JSON.stringify({ model })
    }),

  deleteImageModel: (providerId: string, model: string) =>
    httpJSON<{ ok: boolean }>(
      `/api/settings/image-providers/${encodeURIComponent(providerId)}/models?model=${encodeURIComponent(model)}`,
      { method: "DELETE" }
    ),

  resetImageProvider: (providerId: string) =>
    httpJSON<{ ok: boolean }>(`/api/settings/image-providers/${encodeURIComponent(providerId)}`, {
      method: "DELETE"
    }),

  getHistory: (params?: { capability?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams();
    if (params?.capability) q.set("capability", params.capability);
    if (typeof params?.limit === "number") q.set("limit", String(params.limit));
    if (typeof params?.offset === "number") q.set("offset", String(params.offset));
    const qs = q.toString();
    return httpJSON<{
      items: {
        id: number;
        capability: "image" | "video";
        provider: string;
        model?: string;
        status: string;
        error?: string;
        prompt_preview?: string;
        content_type: string;
        bytes: number;
        url: string;
        created_at: string;
      }[];
    }>(`/api/history${qs ? `?${qs}` : ""}`, { method: "GET" });
  },

  generateImage: (req: ImageGenerateRequest) =>
    httpJSON<ImageGenerateResponse>("/api/images/generate", {
      method: "POST",
      body: JSON.stringify(req)
    }),

  createVideoJob: (req: VideoJobCreateRequest) =>
    httpJSON<VideoJobCreateResponse>("/api/videos/jobs", {
      method: "POST",
      body: JSON.stringify(req)
    }),

  getVideoJob: (jobId: string) =>
    httpJSON<VideoJobGetResponse>(`/api/videos/jobs/${encodeURIComponent(jobId)}`, {
      method: "GET"
    })
};
