import React from "react";
import type { AnimationJobGetResponse } from "../../api_animation";
import { ActionLink, ResultActions } from "../common/ResultActions";

type JobRow = AnimationJobGetResponse & { created_at: number };

export function AnimationResults(props: { jobs: JobRow[]; onDeleteJob: (jobID: string) => void }) {
  const latest = props.jobs[0] || null;
  return (
    <section className="card resultsCard">
      <div className="card__head">
        <h2 className="card__title">生成结果</h2>
        {latest?.video_url ? (
          <div style={{ marginLeft: "auto" }}><ResultActions url={latest.video_url} /></div>
        ) : <div className="badge">{props.jobs.length} jobs</div>}
      </div>
      {latest ? (
        <div className="panel">
          <div className="panel__row"><div className="k">Job</div><div className="v mono">{latest.job_id}</div></div>
          <div className="panel__row"><div className="k">Status</div><div className="v">{latest.status}</div></div>
          <div className="panel__row"><div className="k">Progress</div><div className="v">{latest.completed_segments}/{latest.segment_count || 0}{latest.current_segment ? ` · 当前第 ${latest.current_segment} 段` : ""}</div></div>
          {latest.error ? <div className="alert alert--err">Error: {latest.error}</div> : null}
          {latest.video_url ? <div className="panel__media"><video className="video" controls src={latest.video_url} /></div> : null}
          <div className="animSegments">
            {(latest.segments || []).map((seg) => (
              <div className="animSegments__item" key={`${latest.job_id}_${seg.index}`}>
                <div className="animSegments__head">
                  <div className="animSegments__meta">
                    <div className="mono">#{seg.index + 1}</div>
                    <div className="pill">{seg.status}</div>
                    <div className="muted">{seg.duration_seconds}s</div>
                  </div>
                  {seg.video_url ? <ActionLink href={seg.video_url} compact /> : null}
                </div>
                {seg.video_url ? (
                  <div className="animSegments__video">
                    <video className="video" controls src={seg.video_url} />
                  </div>
                ) : (
                  <div className="animSegments__empty">当前分段还没有可预览视频，状态：{seg.status}</div>
                )}
                <div className="animSegments__prompt">{seg.prompt || "-"}</div>
                {seg.error ? <div className="alert alert--err">Error: {seg.error}</div> : null}
              </div>
            ))}
          </div>
        </div>
      ) : <div className="placeholder"><div className="placeholder__title">结果会显示在这里</div><div className="placeholder__sub">点击 Submit 创建连续动画任务。</div></div>}
      <div className="list">{props.jobs.map((j) => <div className="list__row" key={j.job_id}><div className="mono">{j.job_id}</div><div className="pill">{j.status}</div>{j.video_url ? <ActionLink href={j.video_url} compact /> : <span className="muted">{j.completed_segments}/{j.segment_count || 0}</span>}{j.status !== "succeeded" ? <button className="btn btn--ghost" onClick={() => props.onDeleteJob(j.job_id)} style={{ justifySelf: "end", padding: "6px 10px" }}>删除</button> : <span className="muted">-</span>}</div>)}</div>
    </section>
  );
}
