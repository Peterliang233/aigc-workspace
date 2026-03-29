import type { ModelFormField } from "../../api";

export function fallbackVideoFields(providerID: string, requiresImage: boolean): ModelFormField[] {
  const prompt: ModelFormField = { key: "prompt", label: "Prompt", type: "textarea", required: true, rows: 6 };
  if (providerID === "openai_compatible" || providerID === "bltcy" || providerID === "gpt_best") {
    return [
      prompt,
      ...(requiresImage ? ([{ key: "image", label: "参考图片", type: "image", required: true }] as ModelFormField[]) : []),
      { key: "negative_prompt", label: "Negative Prompt", type: "textarea", rows: 3 },
      { key: "duration_seconds", label: "Duration (sec)", type: "number", placeholder: "例如 5" },
      { key: "aspect_ratio", label: "Aspect Ratio", type: "text", placeholder: "例如 16:9" }
    ];
  }
  return [
    prompt,
    ...(requiresImage ? ([{ key: "image", label: "参考图片", type: "image", required: true }] as ModelFormField[]) : []),
    { key: "negative_prompt", label: "Negative Prompt", type: "textarea", rows: 3 },
    {
      key: "image_size",
      label: "Size",
      type: "select",
      default: "1280x720",
      options: [
        { label: "1280x720 (16:9)", value: "1280x720" },
        { label: "720x1280 (9:16)", value: "720x1280" },
        { label: "960x960 (1:1)", value: "960x960" }
      ]
    },
    { key: "seed", label: "Seed", type: "number", placeholder: "例如 42" }
  ];
}

export function isFieldRequired(fields: ModelFormField[], key: string) {
  key = key.trim();
  return fields.some((f) => String(f.key || "").trim() === key && !!f.required);
}
