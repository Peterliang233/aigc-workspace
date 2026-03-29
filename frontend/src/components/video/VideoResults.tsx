import React from "react";
import type { VideoJobGetResponse } from "../../api";

type JobRow = VideoJobGetResponse & { created_at: number };

export function VideoResults(props: { jobs: JobRow[]; onDeleteJob: (jobID: string) => void }) {
  const { jobs, onDeleteJob } = props;
  const latest = jobs[0] || null;

  return (
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
          {!latest.video_url && latest.status === "succeeded" && <div className="alert">任务成功，但未返回可播放的视频地址。</div>}
        </div>
      ) : (
        <div className="placeholder">
          <div className="placeholder__title">结果会显示在这里</div>
          <div className="placeholder__sub">点击 Submit 创建任务；右侧会自动轮询最新任务状态。</div>
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
            {j.status !== "succeeded" ? (
              <button
                className="btn btn--ghost"
                onClick={() => onDeleteJob(j.job_id)}
                title="删除任务"
                style={{ justifySelf: "end", padding: "6px 10px" }}
              >
                删除
              </button>
            ) : (
              <span className="muted">-</span>
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
