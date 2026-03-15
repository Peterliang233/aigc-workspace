import type { RatioPreset } from "./presets";

export type TargetSize = { w: number; h: number };

export function clampInt(n: number, min: number, max: number) {
  const v = Math.trunc(Number.isFinite(n) ? n : min);
  return Math.min(max, Math.max(min, v));
}

export function targetFromRatio(r: RatioPreset, longEdgePx: number): TargetSize {
  const longEdge = clampInt(longEdgePx, 64, 4096);
  if (r.rw >= r.rh) {
    const w = longEdge;
    const h = clampInt(Math.round((longEdge * r.rh) / r.rw), 1, 4096);
    return { w, h };
  }
  const h = longEdge;
  const w = clampInt(Math.round((longEdge * r.rw) / r.rh), 1, 4096);
  return { w, h };
}

