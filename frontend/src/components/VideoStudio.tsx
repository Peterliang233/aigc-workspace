import React, { useEffect, useMemo, useState } from "react";
import { api, VideoJobGetResponse } from "../api";
import { useVideoMeta } from "../hooks/useVideoMeta";
import { InitImagePicker } from "./video/InitImagePicker";
import { VIDEO_SIZES } from "./video/sizes";
import { VideoResults } from "./video/VideoResults";

type JobRow = VideoJobGetResponse & { created_at: number };

export function VideoStudio() {
  const [prompt, setPrompt] = useState("一段俯拍镜头：城市清晨云海缓慢流动，镜头平滑推进，电影感");
  const [negativePrompt, setNegativePrompt] = useState("");
  const [seed, setSeed] = useState<string>("");
  const [imageSize, setImageSize] = useState("1280x720");
  const [imageUrl, setImageUrl] = useState("");
  const [imageBase64, setImageBase64] = useState("");

  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [jobs, setJobs] = useState<JobRow[]>([]);

  const { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, modelList, useCustom } =
    useVideoMeta();

  const latest = useMemo(() => jobs[0], [jobs]);

  async function onCreateJob() {
    setBusy(true);
    setError(null);
    try {
      const pickedModel = useCustom ? customModel.trim() : model;
      const seedNum = seed.trim() ? Number(seed.trim()) : undefined;
      const req = {
        provider,
        model: pickedModel || undefined,
        prompt,
        negative_prompt: negativePrompt || undefined,
        image_size: imageSize || undefined,
        image: imageBase64 ? imageBase64 : imageUrl.trim() ? imageUrl.trim() : undefined,
        seed: Number.isFinite(seedNum as any) ? (seedNum as number) : undefined
      };
      const res = await api.createVideoJob(req);
      setJobs((prev) => [{ job_id: res.job_id, status: res.status, provider: res.provider, created_at: Date.now() }, ...prev]);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    let t: number | null = null;
    async function tick() {
      if (!latest?.job_id) return;
      if (latest.status === "succeeded" || latest.status === "failed") return;
      try {
        const res = await api.getVideoJob(latest.job_id);
        setJobs((prev) => prev.map((j) => (j.job_id === res.job_id ? { ...j, ...res } : j)));
      } catch {
        // ignore polling errors; user can refresh
      } finally {
        t = window.setTimeout(tick, 1800);
      }
    }
    t = window.setTimeout(tick, 600);
    return () => {
      if (t) window.clearTimeout(t);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [latest?.job_id, latest?.status]);

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head">
          <h2 className="card__title">视频生成</h2>
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
              placeholder="描述你想生成的视频镜头/内容..."
            />
          </label>

          <label className="label">
            Negative Prompt (可选)
            <textarea
              className="textarea"
              value={negativePrompt}
              onChange={(e) => setNegativePrompt(e.target.value)}
              rows={3}
              placeholder="不希望出现的内容..."
            />
          </label>

          <div className="row2">
            <label className="label">
              Size
              <select className="input" value={imageSize} onChange={(e) => setImageSize(e.target.value)}>
                {VIDEO_SIZES.map((s) => (
                  <option key={s.key} value={s.key}>
                    {s.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="label">
              Seed (可选)
              <input className="input" value={seed} onChange={(e) => setSeed(e.target.value)} placeholder="例如 42" />
            </label>
          </div>

          <InitImagePicker
            imageUrl={imageUrl}
            onImageUrl={setImageUrl}
            imageBase64={imageBase64}
            onImageBase64={setImageBase64}
            disabled={busy}
          />

          <button className="btn" disabled={busy || !prompt.trim()} onClick={onCreateJob}>
            {busy ? "Submitting..." : "Submit"}
          </button>
        </div>

        {error && <div className="alert alert--err">Error: {error}</div>}
      </section>

      <VideoResults jobs={jobs} />
    </div>
  );
}
