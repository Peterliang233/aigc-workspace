export type ImageGenerateRequest = {
  prompt: string;
  size?: string;
  n?: number;
};

export type ImageGenerateResponse = {
  image_urls: string[];
  provider: string;
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

