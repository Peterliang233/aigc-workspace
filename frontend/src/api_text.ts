import type { ProviderMetaResponse } from "./api";

export type TextGenerateRequest = {
  provider?: string;
  model?: string;
  prompt: string;
  system_prompt?: string;
  temperature?: number;
  max_tokens?: number;
};

export type TextGenerateResponse = {
  text: string;
  provider: string;
  model?: string;
};

async function httpJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: { "Content-Type": "application/json", ...(init?.headers || {}) }
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(typeof data?.error === "string" ? data.error : `HTTP ${res.status}`);
  return data as T;
}

export const textApi = {
  getMeta: () => httpJSON<ProviderMetaResponse>("/api/meta/texts", { method: "GET" }),
  generate: (req: TextGenerateRequest) =>
    httpJSON<TextGenerateResponse>("/api/texts/generate", { method: "POST", body: JSON.stringify(req) })
};
