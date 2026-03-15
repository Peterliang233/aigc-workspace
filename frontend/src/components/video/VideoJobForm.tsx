import React, { useMemo, useState } from "react";
import type { ModelFormField } from "../../api";
import { useVideoMeta } from "../../hooks/useVideoMeta";
import { useGeneration } from "../../state/generation";
import { ModelFields } from "../form/ModelFields";
import { useApplyFieldDefaults } from "../form/useApplyFieldDefaults";
import { InitImageField } from "./InitImageField";
import { fallbackVideoFields, isFieldRequired } from "./videoFormDefaults";

export function VideoJobForm(props: { latestStatus?: string }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    metaLoading,
    providers,
    provider,
    setProvider,
    model,
    setModel,
    customModel,
    setCustomModel,
    models,
    modelList,
    selectedModelMeta,
    useCustom
  } = useVideoMeta();

  const { createVideoJob } = useGeneration();

  const schemaRequiresImage = !!selectedModelMeta?.requires_image || !!selectedModelMeta?.form?.requires_image;
  const customRequiresImage = useCustom && /I2V|IMG2VID|IMAGE2VIDEO/i.test(customModel.trim());
  const requiresImage = schemaRequiresImage || customRequiresImage;

  const fields = useMemo(() => {
    const sf = selectedModelMeta?.form?.fields || [];
    if (Array.isArray(sf) && sf.length > 0) {
      const hasPrompt = sf.some((f) => String(f.key || "").trim() === "prompt");
      if (hasPrompt) return sf;
      return [{ key: "prompt", label: "Prompt", type: "textarea", required: true, rows: 6 }, ...sf];
    }
    return fallbackVideoFields(provider, requiresImage);
  }, [selectedModelMeta, provider, requiresImage]);

  const [values, setValues] = useState<Record<string, string>>(() => ({
    prompt: "一段俯拍镜头：城市清晨云海缓慢流动，镜头平滑推进，电影感"
  }));

  useApplyFieldDefaults(fields, setValues, [provider, model, customModel]);

  const missing = useMemo(() => {
    const out: string[] = [];
    if (useCustom && !customModel.trim()) out.push("model");
    for (const f of fields) {
      const k = String(f.key || "").trim();
      if (!k || !f.required) continue;
      if (!String(values[k] || "").trim()) out.push(k);
    }
    // keep legacy flag behavior for models that only set requires_image
    if (requiresImage && !String(values["image"] || "").trim() && !isFieldRequired(fields, "image")) out.push("image");
    return out;
  }, [fields, values, requiresImage]);

  const canSubmit = !!String(values["prompt"] || "").trim() && missing.length === 0;

  async function onSubmit() {
    setBusy(true);
    setError(null);
    try {
      const pickedModel = useCustom ? customModel.trim() : model;
      const seedNum = String(values["seed"] || "").trim() ? Number(String(values["seed"]).trim()) : undefined;
      const durNum = String(values["duration_seconds"] || "").trim()
        ? Number(String(values["duration_seconds"]).trim())
        : undefined;

      await createVideoJob({
        provider,
        model: pickedModel || undefined,
        prompt: String(values["prompt"] || ""),
        negative_prompt: String(values["negative_prompt"] || "").trim() || undefined,
        image_size: String(values["image_size"] || "").trim() || undefined,
        image: String(values["image"] || "").trim() || undefined,
        seed: Number.isFinite(seedNum as any) ? (seedNum as number) : undefined,
        duration_seconds: Number.isFinite(durNum as any) ? (durNum as number) : undefined,
        aspect_ratio: String(values["aspect_ratio"] || "").trim() || undefined
      });
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">视频生成</h2>
      </div>

      <div className="form">
        <div className="row2">
          <label className="label">
            平台
            <select className="input" value={provider} onChange={(e) => setProvider(e.target.value)} disabled={metaLoading}>
              {providers.map((p) => (
                <option key={p.id} value={p.id} disabled={!p.configured}>
                  {p.label}
                  {!p.configured ? " (未配置)" : ""}
                </option>
              ))}
            </select>
          </label>

          <label className="label">
            模型
            {modelList.length > 0 ? (
              <select className="input" value={model} onChange={(e) => setModel(e.target.value)}>
                {(models || []).map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.label ? `${m.label} (${m.id})` : m.id}
                  </option>
                ))}
                <option value="__custom__">自定义...</option>
              </select>
            ) : (
              <input
                className="input"
                value={customModel}
                onChange={(e) => setCustomModel(e.target.value)}
                placeholder="输入模型名称"
              />
            )}
          </label>
        </div>

        {modelList.length > 0 && useCustom && (
          <label className="label">
            自定义模型
            <input
              className="input"
              value={customModel}
              onChange={(e) => setCustomModel(e.target.value)}
              placeholder="输入模型名称"
            />
          </label>
        )}

        <ModelFields
          fields={fields}
          values={values}
          onChange={(k, v) => setValues((prev) => ({ ...prev, [k]: v }))}
          disabled={busy}
          custom={(f) => {
            if (String(f.type || "").trim() !== "image") return null;
            const k = String(f.key || "").trim();
            if (!k) return null;
            return (
              <InitImageField
                key={k}
                value={values[k] || ""}
                onChange={(v) => setValues((prev) => ({ ...prev, [k]: v }))}
                disabled={busy}
                required={!!f.required || requiresImage}
              />
            );
          }}
        />

        <button className="btn" disabled={busy || !canSubmit} onClick={onSubmit}>
          {busy ? "Submitting..." : "Submit"}
        </button>
      </div>

      {error && <div className="alert alert--err">Error: {error}</div>}
      {missing.length > 0 ? <div className="alert">缺少必填字段：{missing.join(", ")}</div> : null}
      {props.latestStatus && props.latestStatus !== "succeeded" && props.latestStatus !== "failed" ? (
        <div className="alert">最新任务正在生成中，切换 tab 不会中断。</div>
      ) : null}
    </section>
  );
}
