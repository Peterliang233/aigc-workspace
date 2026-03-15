import React, { useState } from "react";
import { api } from "../api";

export function ImageStudio() {
  const [prompt, setPrompt] = useState("一只在雨夜霓虹街头散步的柴犬，电影感，高对比，35mm");
  const [size, setSize] = useState("1024x1024");
  const [n, setN] = useState(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [urls, setUrls] = useState<string[]>([]);
  const [selected, setSelected] = useState<string | null>(null);
  const [provider, setProvider] = useState<string>("-");

  async function onGenerate() {
    setBusy(true);
    setError(null);
    try {
      const res = await api.generateImage({ prompt, size, n });
      const next = res.image_urls || [];
      setUrls(next);
      setSelected(next[0] || null);
      setProvider(res.provider || "-");
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
          <div className="badge">provider: {provider}</div>
        </div>

        <div className="form">
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
              <div className="thumb__cap">{u.startsWith("/static/") ? "local" : "remote"}</div>
            </button>
          ))}
        </div>
      </section>
    </div>
  );
}
