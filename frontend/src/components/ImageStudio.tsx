import React, { useMemo, useState } from "react";
import { useImageMeta } from "../hooks/useImageMeta";
import { useGeneration } from "../state/generation";

export function ImageStudio() {
  const [prompt, setPrompt] = useState("一只在雨夜霓虹街头散步的柴犬，电影感，高对比，35mm");
  const [size, setSize] = useState("1024x1024");

  const { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, modelList, useCustom } =
    useImageMeta();

  const { state, startImage } = useGeneration();
  const latest = useMemo(() => state.images[0] || null, [state.images]);
  const busy = state.images.some((t) => t.status === "running");
  const url = latest?.image_url || null;

  function onGenerate() {
    const pickedModel = useCustom ? customModel.trim() : model;
    startImage({
      provider,
      model: pickedModel || undefined,
      prompt,
      size
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
                  {modelList.map((m) => (
                    <option key={m} value={m}>
                      {m}
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

          <label className="label">
            Prompt
            <textarea
              className="textarea"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              rows={6}
              placeholder="描述你想生成的画面..."
            />
          </label>

          <div className="row row--image">
            <label className="label">
              Size
              <input className="input" value={size} onChange={(e) => setSize(e.target.value)} />
            </label>
            <button className="btn" disabled={busy} onClick={onGenerate}>
              {busy ? "Generating..." : "Generate"}
            </button>
          </div>
        </div>

        {latest?.status === "failed" && <div className="alert alert--err">Error: {latest.error || "failed"}</div>}
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
            <img className="preview__img" src={url} alt={prompt} />
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
