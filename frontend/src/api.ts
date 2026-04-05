export type ImageGenerateRequest = {
  provider?: string;
  model?: string;
  prompt: string;
  size?: string;
  n?: number;
  negative_prompt?: string;
  aspect_ratio?: string;
  image?: string[];
  reference_urls?: string[];
  seed?: number;
  strength?: number;
  style?: string;
};

export type ImageGenerateResponse = {
  image_urls: string[];
  provider: string;
  model?: string;
};

export type ModelFormFieldOption = { label?: string; value: string };

export type ModelFormField = {
  key: string;
  label?: string;
  type: string; // text|textarea|select|number|image
  required?: boolean;
  placeholder?: string;
  default?: any;
  options?: ModelFormFieldOption[];
  rows?: number;
};

export type ModelForm = { requires_image?: boolean; fields?: ModelFormField[] };

export type ProviderModelMeta = {
  id: string;
  label?: string;
  requires_image?: boolean;
  form?: ModelForm;
};

export type ProviderMetaResponse = {
  default_provider: string;
  providers: {
    id: string;
    label: string;
    configured: boolean;
    models: ProviderModelMeta[];
  }[];
};

export type VideoJobCreateRequest = {
  provider?: string;
  model?: string;
  prompt: string;
  duration_seconds?: number;
  aspect_ratio?: string;
  image_size?: string;
  negative_prompt?: string;
  image?: string;
  seed?: number;
  extra?: Record<string, unknown>;
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

export type VideoMetaResponse = {
  default_provider: string;
  providers: {
    id: string;
    label: string;
    configured: boolean;
    models: ProviderModelMeta[];
  }[];
};

export type AudioGenerateRequest = {
  provider?: string;
  model?: string;
  input: string;
  voice?: string;
  response_format?: string;
  speed?: number;
};

export type AudioGenerateResponse = {
  audio_url: string;
  provider: string;
  model?: string;
  content_type?: string;
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
    httpJSON<ProviderMetaResponse>("/api/meta/images", { method: "GET" }),

  getVideoMeta: () => httpJSON<VideoMetaResponse>("/api/meta/videos", { method: "GET" }),
  getAudioMeta: () => httpJSON<ProviderMetaResponse>("/api/meta/audios", { method: "GET" }),

  getHistory: (params?: { capability?: string; q?: string; limit?: number; offset?: number; page?: number; page_size?: number }) => {
    const q = new URLSearchParams();
    if (params?.capability) q.set("capability", params.capability);
    if (params?.q) q.set("q", params.q);
    if (typeof params?.limit === "number") q.set("limit", String(params.limit));
    if (typeof params?.offset === "number") q.set("offset", String(params.offset));
    if (typeof params?.page === "number") q.set("page", String(params.page));
    if (typeof params?.page_size === "number") q.set("page_size", String(params.page_size));
    const qs = q.toString();
    return httpJSON<{
      items: {
        id: number;
        capability: "image" | "video" | "audio";
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
      total?: number;
      page?: number;
      page_size?: number;
    }>(`/api/history${qs ? `?${qs}` : ""}`, { method: "GET" });
  },

  generateImage: (req: ImageGenerateRequest) =>
    httpJSON<ImageGenerateResponse>("/api/images/generate", {
      method: "POST",
      body: JSON.stringify(req)
    }),

  generateAudio: (req: AudioGenerateRequest) =>
    httpJSON<AudioGenerateResponse>("/api/audios/generate", {
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
    }),

  deleteHistory: (id: number) =>
    httpJSON<{ ok: boolean; id: number }>(`/api/history/${encodeURIComponent(String(id))}`, {
      method: "DELETE"
    })
};
