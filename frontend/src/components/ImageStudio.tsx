import React, { useEffect, useMemo, useState } from "react";
import { api } from "../api";

export function ImageStudio() {
  const [prompt, setPrompt] = useState("一只在雨夜霓虹街头散步的柴犬，电影感，高对比，35mm");
  const [size, setSize] = useState("1024x1024");
  const [n, setN] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [urls, setUrls] = useState<string[]>([]);
  const [selected, setSelected] = useState<string | null>(null);

  const [metaLoading, setMetaLoading] = useState(false);
  const [providers, setProviders] = useState<
    { id: string; label: string; configured: boolean; models: string[] }[]
  >([]);
  const [provider, setProvider] = useState<string>(() => localStorage.getItem("aigc_image_provider") || "");
  const [model, setModel] = useState<string>(() => localStorage.getItem("aigc_image_model") || "");
  const [customModel, setCustomModel] = useState<string>(() => localStorage.getItem("aigc_image_custom_model") || "");

  const providerInfo = useMemo(
    () => providers.find((p) => p.id === provider) || null,
    [providers, provider]
  );
  const modelList = providerInfo?.models || [];
  const useCustom = model === "__custom__" || modelList.length === 0;

  useEffect(() => {
    let mounted = true;
    async function load() {
      setMetaLoading(true);
      try {
        const res = await api.getImageMeta();
        if (!mounted) return;
        const list = (res.providers || []).slice().sort((a, b) => a.label.localeCompare(b.label));
        setProviders(list);

        // pick default provider if none selected or selected is missing
        const preferred = provider || res.default_provider || "mock";
        const hasPreferred = list.some((p) => p.id === preferred && p.configured);
        const fallback = list.find((p) => p.configured)?.id || "mock";
        const nextProvider = hasPreferred ? preferred : fallback;
        setProvider(nextProvider);

        // pick default model if empty
        const pi = list.find((p) => p.id === nextProvider);
        const models = pi?.models || [];
        if (!model) {
          setModel(models[0] || "__custom__");
        }
      } catch (e: any) {
        // if meta fails, fall back to mock with no model
        if (mounted) {
          setProviders([{ id: "mock", label: "Mock(联调)", configured: true, models: [] }]);
          if (!provider) setProvider("mock");
          if (!model) setModel("__custom__");
        }
      } finally {
        if (mounted) setMetaLoading(false);
      }
    }
    load();
    return () => {
      mounted = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (provider) localStorage.setItem("aigc_image_provider", provider);
  }, [provider]);
  useEffect(() => {
    if (model) localStorage.setItem("aigc_image_model", model);
  }, [model]);
  useEffect(() => {
    if (customModel) localStorage.setItem("aigc_image_custom_model", customModel);
  }, [customModel]);

  // When provider changes, reset model to first model (or custom) if current model isn't compatible.
  useEffect(() => {
    if (!providerInfo) return;
    const models = providerInfo.models || [];
    if (models.length === 0) {
      if (model !== "__custom__") setModel("__custom__");
      return;
    }
    if (model === "__custom__") return;
    if (!models.includes(model)) {
      setModel(models[0]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [provider]);

  async function onGenerate() {
    setBusy(true);
    setError(null);
    try {
      const pickedModel = useCustom ? customModel.trim() : model;
      const res = await api.generateImage({
        provider,
        model: pickedModel || undefined,
        prompt,
        size,
        n
      });
      const next = res.image_urls || [];
      setUrls(next);
      setSelected(next[0] || null);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
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

          <div className="row">
            <label className="label">
              Size
              <input className="input" value={size} onChange={(e) => setSize(e.target.value)} />
            </label>
            <label className="label">
              N
              <input
                className="input"
                type="number"
                min={1}
                max={4}
                value={n}
                onChange={(e) => setN(Number(e.target.value))}
              />
            </label>
            <button className="btn" disabled={busy} onClick={onGenerate}>
              {busy ? "Generating..." : "Generate"}
            </button>
          </div>
        </div>

        {error && <div className="alert alert--err">Error: {error}</div>}
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">生成结果</h2>
          <div className="badge">{urls.length} images</div>
        </div>

        {selected ? (
          <a className="preview" href={selected} target="_blank" rel="noreferrer">
            <img className="preview__img" src={selected} alt={prompt} />
          </a>
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">结果会显示在这里</div>
            <div className="placeholder__sub">点击 Generate 生成图片；生成后可点击缩略图切换预览。</div>
          </div>
        )}

        <div className="thumbs">
          {urls.map((u) => (
            <button
              className={u === selected ? "thumb thumb--active" : "thumb"}
              key={u}
              onClick={() => setSelected(u)}
              title="Select"
            >
              <img className="thumb__img" src={u} alt={prompt} />
            </button>
          ))}
        </div>
      </section>
    </div>
  );
}
