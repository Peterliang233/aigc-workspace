export type Preset = { key: string; w: number; h: number; label: string };
export type Mode = "cover" | "contain";
export type RatioPreset = { key: string; rw: number; rh: number; label: string };

export const PRESETS: Preset[] = [
  { key: "1024x1024", w: 1024, h: 1024, label: "1024x1024 (1:1)" },
  { key: "1024x1536", w: 1024, h: 1536, label: "1024x1536 (2:3)" },
  { key: "1536x1024", w: 1536, h: 1024, label: "1536x1024 (3:2)" },
  { key: "1280x720", w: 1280, h: 720, label: "1280x720 (16:9)" },
  { key: "720x1280", w: 720, h: 1280, label: "720x1280 (9:16)" }
];

export const RATIOS: RatioPreset[] = [
  { key: "1:1", rw: 1, rh: 1, label: "1:1" },
  { key: "4:3", rw: 4, rh: 3, label: "4:3" },
  { key: "3:4", rw: 3, rh: 4, label: "3:4" },
  { key: "3:2", rw: 3, rh: 2, label: "3:2" },
  { key: "2:3", rw: 2, rh: 3, label: "2:3" },
  { key: "16:9", rw: 16, rh: 9, label: "16:9" },
  { key: "9:16", rw: 9, rh: 16, label: "9:16" }
];
