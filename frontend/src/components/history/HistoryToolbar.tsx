import React from "react";

export function HistoryToolbar(props: {
  capability: "all" | "image" | "video" | "audio";
  setCapability: (v: "all" | "image" | "video" | "audio") => void;
  qInput: string;
  setQInput: (v: string) => void;
  onSearch: () => void;
  pageSize: number;
  setPageSize: (v: number) => void;
  total: number;
  page: number;
  pages: number;
  busy: boolean;
  onRefresh: () => void;
  onPrev: () => void;
  onNext: () => void;
}) {
  const p = props;
  return (
    <>
      <div className="card__head">
        <h2 className="card__title">历史</h2>
        <div className="chips">
          <button className={p.capability === "all" ? "chip" : "chip chip--ghost"} onClick={() => p.setCapability("all")} disabled={p.busy}>
            <span className="chip__text">全部</span>
          </button>
          <button className={p.capability === "image" ? "chip" : "chip chip--ghost"} onClick={() => p.setCapability("image")} disabled={p.busy}>
            <span className="chip__text">图片</span>
          </button>
          <button className={p.capability === "video" ? "chip" : "chip chip--ghost"} onClick={() => p.setCapability("video")} disabled={p.busy}>
            <span className="chip__text">视频</span>
          </button>
          <button className={p.capability === "audio" ? "chip" : "chip chip--ghost"} onClick={() => p.setCapability("audio")} disabled={p.busy}>
            <span className="chip__text">音频</span>
          </button>
          <button className="btn btn--ghost" onClick={p.onRefresh} disabled={p.busy} title="Refresh">
            {p.busy ? "Loading..." : "Refresh"}
          </button>
        </div>
      </div>

      <div className="row2">
        <label className="label">
          查询
          <input
            className="input"
            value={p.qInput}
            onChange={(e) => p.setQInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") p.onSearch();
            }}
            placeholder="按ID / 平台 / 模型 / 提示词搜索"
            disabled={p.busy}
          />
        </label>
        <label className="label">
          每页
          <select className="input" value={String(p.pageSize)} onChange={(e) => p.setPageSize(Number(e.target.value) || 5)} disabled={p.busy}>
            <option value="10">5</option>
            <option value="20">10</option>
            <option value="50">20</option>
          </select>
        </label>
      </div>

      <div className="chips" style={{ marginBottom: 10 }}>
        <button className="btn btn--ghost" onClick={p.onSearch} disabled={p.busy}>
          查询
        </button>
        <span className="muted">共 {p.total} 条</span>
        <span className="muted">第 {p.page} / {p.pages} 页</span>
        <button className="btn btn--ghost" disabled={p.busy || p.page <= 1} onClick={p.onPrev}>
          上一页
        </button>
        <button className="btn btn--ghost" disabled={p.busy || p.page >= p.pages} onClick={p.onNext}>
          下一页
        </button>
      </div>
    </>
  );
}
