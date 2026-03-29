import React, { useMemo, useState } from "react";
import { useImageMeta } from "../hooks/useImageMeta";
import { useGeneration } from "../state/generation";
import { ModelFields } from "./form/ModelFields";
import { useApplyFieldDefaults } from "./form/useApplyFieldDefaults";
import type { ModelFormField } from "../api";

export function ImageStudio() {
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
  } = useImageMeta();

  const { state, startImage } = useGeneration();
  const latest = useMemo(() => state.images[0] || null, [state.images]);
  const busy = state.images.some((t) => t.status === "running");
  const url = latest?.image_url || null;

  const fields: ModelFormField[] = useMemo(() => {
    const sf = selectedModelMeta?.form?.fields || [];
    if (Array.isArray(sf) && sf.length > 0) {
      const hasPrompt = sf.some((f) => String(f.key || "").trim() === "prompt");
      if (hasPrompt) return sf;
      return [{ key: "prompt", label: "Prompt", type: "textarea", required: true, rows: 6 }, ...sf];
    }
    return [
      { key: "prompt", label: "Prompt", type: "textarea", required: true, rows: 6 },
      { key: "size", label: "Size", type: "text", placeholder: "1024x1024" }
    ];
  }, [selectedModelMeta]);

  const [values, setValues] = useState<Record<string, string>>(() => ({
    prompt: "一只在雨夜霓虹街头散步的柴犬，电影感，高对比，35mm",
    size: "1024x1024"
  }));

  useApplyFieldDefaults(fields, setValues, [provider, model, customModel]);

  const missing = useMemo(() => {
    const out: string[] = [];
    for (const f of fields) {
      if (!f.required) continue;
      const k = String(f.key || "").trim();
      if (!k) continue;
      if (!String(values[k] || "").trim()) out.push(k);
    }
    return out;
  }, [fields, values]);

  function onGenerate() {
    const pickedModel = useCustom ? customModel.trim() : model;
    const seedNum = String(values["seed"] || "").trim() ? Number(String(values["seed"]).trim()) : undefined;
    startImage({
      provider,
      model: pickedModel || undefined,
      prompt: String(values["prompt"] || ""),
      size: String(values["size"] || "").trim() || undefined,
      negative_prompt: String(values["negative_prompt"] || "").trim() || undefined,
      aspect_ratio: String(values["aspect_ratio"] || "").trim() || undefined,
      seed: Number.isFinite(seedNum as any) ? (seedNum as number) : undefined
    });
  }

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head">
          <h2 className="card__title">图片生成</h2>
        </div>

        <div className="form">
          <div className="row2">
            <label className="label">
              平台
              <select
                className="input"
                value={provider}
                onChange={(e) => setProvider(e.target.value)}
                disabled={metaLoading}
              >
                {providers.map((p) => (
                  <option key={p.id} value={p.id}>
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
          />

          <button className="btn" disabled={busy || missing.length > 0} onClick={onGenerate}>
            {busy ? "Generating..." : "Generate"}
          </button>
        </div>

        {latest?.status === "failed" && <div className="alert alert--err">Error: {latest.error || "failed"}</div>}
        {missing.length > 0 ? <div className="alert">缺少必填字段：{missing.join(", ")}</div> : null}
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">生成结果</h2>
          {url ? (
            <div className="chips" style={{ marginLeft: "auto" }}>
              {url.startsWith("/api/assets/") ? (
                <a className="btn btn--ghost" href={`${url}?download=1`} title="下载图片">
                  下载
                </a>
              ) : (
                <a className="btn btn--ghost" href={url} download title="下载图片">
                  下载
                </a>
              )}
            </div>
          ) : null}
        </div>

        {url ? (
          <a className="preview" href={url} target="_blank" rel="noreferrer" title="Open original">
            <img className="preview__img" src={url} alt={values["prompt"] || ""} />
          </a>
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">{busy ? "生成中..." : "结果会显示在这里"}</div>
            <div className="placeholder__sub">
              {busy ? "切换 tab 不会中断；完成后会自动显示。" : "点击 Generate 生成图片。"}
            </div>
          </div>
        )}
      </section>
    </div>
  );
}
