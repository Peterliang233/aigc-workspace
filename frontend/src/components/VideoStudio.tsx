import React, { useEffect, useMemo, useState } from "react";
import { api, VideoJobGetResponse } from "../api";

type JobRow = VideoJobGetResponse & { created_at: number };

export function VideoStudio() {
  const [prompt, setPrompt] = useState("一段俯拍镜头：城市清晨云海缓慢流动，镜头平滑推进，电影感");
  const [duration, setDuration] = useState(5);
  const [aspect, setAspect] = useState("16:9");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [jobs, setJobs] = useState<JobRow[]>([]);

  const latest = useMemo(() => jobs[0], [jobs]);

  async function onCreateJob() {
    setBusy(true);
    setError(null);
    try {
      const res = await api.createVideoJob({
        prompt,
        duration_seconds: duration,
        aspect_ratio: aspect
      });
      setJobs((prev) => [
        {
          job_id: res.job_id,
          status: res.status,
          provider: res.provider,
          created_at: Date.now()
        },
        ...prev
      ]);
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
        setJobs((prev) =>
          prev.map((j) => (j.job_id === res.job_id ? { ...j, ...res } : j))
        );
      } catch {
        // ignore polling errors; user can refresh
      } finally {
        t = window.setTimeout(tick, 1500);
      }
    }
    t = window.setTimeout(tick, 500);
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

          <div className="row">
            <label className="label">
              Duration(s)
              <input
                className="input"
                type="number"
                min={1}
                max={20}
                value={duration}
                onChange={(e) => setDuration(Number(e.target.value))}
              />
            </label>
            <label className="label">
              Aspect
              <input className="input" value={aspect} onChange={(e) => setAspect(e.target.value)} />
            </label>
            <button className="btn" disabled={busy} onClick={onCreateJob}>
              {busy ? "Creating..." : "Create Job"}
            </button>
          </div>
        </div>

        {error && <div className="alert alert--err">Error: {error}</div>}
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">生成结果</h2>
          <div className="badge">{jobs.length} jobs</div>
        </div>

        {latest ? (
          <div className="panel">
            <div className="panel__row">
              <div className="k">Job</div>
              <div className="v mono">{latest.job_id}</div>
            </div>
            <div className="panel__row">
              <div className="k">Status</div>
              <div className="v">{latest.status}</div>
            </div>
            {latest.error && (
              <div className="panel__row">
                <div className="k">Error</div>
                <div className="v">{latest.error}</div>
              </div>
            )}
            {latest.video_url && (
              <div className="panel__media">
                <video className="video" controls src={latest.video_url} />
              </div>
            )}
            {!latest.video_url && latest.status === "succeeded" && (
              <div className="alert">任务成功，但未返回可播放的视频地址。</div>
            )}
          </div>
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">结果会显示在这里</div>
            <div className="placeholder__sub">点击 Create Job 创建任务；右侧会自动轮询最新任务状态。</div>
          </div>
        )}

        <div className="list">
          {jobs.map((j) => (
            <div className="list__row" key={j.job_id}>
              <div className="mono">{j.job_id}</div>
              <div className="pill">{j.status}</div>
              {j.video_url ? (
                <a className="link" href={j.video_url} target="_blank" rel="noreferrer">
                  open
                </a>
              ) : (
                <span className="muted">-</span>
              )}
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
