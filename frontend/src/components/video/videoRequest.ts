import type { ModelFormField } from "../../api";

const coreVideoFields = new Set([
  "prompt",
  "negative_prompt",
  "image_size",
  "image",
  "seed",
  "duration_seconds",
  "aspect_ratio"
]);

export function pickVideoExtra(fields: ModelFormField[], values: Record<string, string>) {
  const extra: Record<string, unknown> = {};
  for (const field of fields) {
    const key = String(field.key || "").trim();
    if (!key || coreVideoFields.has(key)) continue;
    const raw = values[key];
    if (typeof raw !== "string" || raw.trim() === "") continue;
    const lowered = raw.trim().toLowerCase();
    if (field.type === "number") {
      const num = Number(raw.trim());
      extra[key] = Number.isFinite(num) ? num : raw;
      continue;
    }
    if (field.type === "select" && (lowered === "true" || lowered === "false")) {
      extra[key] = lowered === "true";
      continue;
    }
    extra[key] = raw;
  }
  return Object.keys(extra).length > 0 ? extra : undefined;
}
