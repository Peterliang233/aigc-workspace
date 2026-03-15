import React, { useEffect, useMemo, useState } from "react";
import { api } from "../api";

type Item = {
  id: number;
  capability: "image" | "video";
  provider: string;
  model?: string;
  status: string;
  error?: string;
  prompt_preview?: string;
  content_type: string;
  bytes: number;
  url: string;
  created_at: string;
};

function fmtBytes(n: number) {
  if (!Number.isFinite(n) || n <= 0) return "-";
  const units = ["B", "KB", "MB", "GB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(v < 10 && i > 0 ? 1 : 0)} ${units[i]}`;
}

export function HistoryStudio() {
  const [capability, setCapability] = useState<"all" | "image" | "video">("all");
  const [items, setItems] = useState<Item[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<number | null>(null);

  const selected = useMemo(
    () => items.find((x) => x.id === selectedId) || null,
    [items, selectedId]
  );

  async function load() {
    setBusy(true);
    setError(null);
    try {
      const res = await api.getHistory({
        capability: capability === "all" ? undefined : capability,
        limit: 50,
        offset: 0
      });
      setItems(res.items || []);
      const first = (res.items || [])[0];
      if (!selectedId && first) setSelectedId(first.id);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [capability]);

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head">
          <h2 className="card__title">历史</h2>
          <div className="chips">
            <button
              className={capability === "all" ? "chip" : "chip chip--ghost"}
              onClick={() => setCapability("all")}
              disabled={busy}
            >
              <span className="chip__text">全部</span>
            </button>
            <button
              className={capability === "image" ? "chip" : "chip chip--ghost"}
              onClick={() => setCapability("image")}
              disabled={busy}
            >
              <span className="chip__text">图片</span>
            </button>
            <button
              className={capability === "video" ? "chip" : "chip chip--ghost"}
              onClick={() => setCapability("video")}
              disabled={busy}
            >
              <span className="chip__text">视频</span>
            </button>
            <button className="btn btn--ghost" onClick={load} disabled={busy} title="Refresh">
              {busy ? "Loading..." : "Refresh"}
            </button>
          </div>
        </div>

        {error && <div className="alert alert--err">Error: {error}</div>}

        <div className="list">
          {items.map((it) => (
            <button
              key={it.id}
              className={it.id === selectedId ? "hrow hrow--active" : "hrow"}
              onClick={() => setSelectedId(it.id)}
              title="Select"
            >
              <div className="hrow__top">
                <div className="mono">#{it.id}</div>
                <div className="pill">{it.capability}</div>
              </div>
              <div className="hrow__mid">
                <div className="muted">{it.provider}</div>
                <div className="muted">{fmtBytes(it.bytes)}</div>
              </div>
              <div className="hrow__bot">
                <div className="hrow__prompt">{it.prompt_preview || ""}</div>
                <div className="hrow__time mono">{it.created_at}</div>
              </div>
            </button>
          ))}
          {items.length === 0 && !busy && (
            <div className="placeholder">
              <div className="placeholder__title">暂无记录</div>
              <div className="placeholder__sub">生成图片或视频后，会自动保存到 MinIO 并出现在这里。</div>
            </div>
          )}
        </div>
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">结果预览</h2>
          {selected ? (
            <a className="link" href={selected.url} target="_blank" rel="noreferrer">
              open
            </a>
          ) : (
            <span className="muted">-</span>
          )}
        </div>

        {selected ? (
          selected.capability === "video" ? (
            <div className="panel__media">
              <video className="video" controls src={selected.url} />
            </div>
          ) : (
            <a className="preview" href={selected.url} target="_blank" rel="noreferrer">
              <img className="preview__img" src={selected.url} alt={selected.prompt_preview || ""} />
            </a>
          )
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">选择一条记录预览</div>
            <div className="placeholder__sub">左侧列表选择图片或视频。</div>
          </div>
        )}

        {selected?.error && <div className="alert alert--err">Error: {selected.error}</div>}
      </section>
    </div>
  );
}

