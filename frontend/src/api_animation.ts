export type AnimationJobCreateRequest = {
  provider?: string;
  model?: string;
  planner_model?: string;
  prompt: string;
  duration_seconds: number;
  aspect_ratio?: string;
  lead_image?: string;
  seed?: number;
};

export type AnimationSegment = {
  index: number;
  status: string;
  duration_seconds: number;
  prompt?: string;
  continuity?: string;
  source_job_id?: string;
  video_url?: string;
  last_frame_ready?: boolean;
  error?: string;
};

export type AnimationJobCreateResponse = {
  job_id: string;
  status: string;
  provider: string;
  model?: string;
  duration_seconds: number;
};

export type AnimationJobGetResponse = AnimationJobCreateResponse & {
  prompt?: string;
  planner_status?: string;
  planner_model?: string;
  planner_error?: string;
  segment_count: number;
  completed_segments: number;
  current_segment: number;
  video_url?: string;
  error?: string;
  segments?: AnimationSegment[];
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

export const animationApi = {
  createJob: (req: AnimationJobCreateRequest) =>
    httpJSON<AnimationJobCreateResponse>("/api/animations/jobs", {
      method: "POST",
      body: JSON.stringify(req)
    }),

  getJob: (jobId: string) =>
    httpJSON<AnimationJobGetResponse>(`/api/animations/jobs/${encodeURIComponent(jobId)}`, {
      method: "GET"
    })
};
