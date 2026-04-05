import React, { useEffect, useMemo, useState } from "react";
import { api } from "../api";
import { HistoryRow } from "./history/HistoryRow";
import { HistoryToolbar } from "./history/HistoryToolbar";
type Item = {
  id: number;
  capability: "image" | "video" | "audio";
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
export function HistoryStudio() {
  const [capability, setCapability] = useState<"all" | "image" | "video" | "audio">("all");
  const [q, setQ] = useState("");
  const [qInput, setQInput] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);
  const [total, setTotal] = useState(0);
  const [items, setItems] = useState<Item[]>([]);
  const [busy, setBusy] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const selected = useMemo(() => items.find((x) => x.id === selectedId) || null, [items, selectedId]);
  const pages = Math.max(1, Math.ceil(total / Math.max(1, pageSize)));
  async function load(targetPage = page) {
    setBusy(true);
    setError(null);
    try {
      const res = await api.getHistory({
        capability: capability === "all" ? undefined : capability,
        q: q || undefined,
        page: targetPage,
        page_size: pageSize
      });
      const list = res.items || [];
      setItems(list);
      setTotal(Number(res.total || 0));
      setPage(Number(res.page || targetPage));
      const first = (res.items || [])[0];
      if (!first) {
        setSelectedId(null);
      } else if (!list.some((x) => x.id === selectedId)) {
        setSelectedId(first.id);
      }
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    void load(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [capability, q, pageSize]);

  async function onDeleteItem(it: Item) {
    const ok = window.confirm(`确定永久删除记录 #${it.id} 吗？删除后不可恢复。`);
    if (!ok) return;
    setDeletingId(it.id);
    setError(null);
    try {
      await api.deleteHistory(it.id);
      const nextPage = page > 1 && items.length <= 1 ? page - 1 : page;
      await load(nextPage);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setDeletingId(null);
    }
  }
  return (
    <div className="workspace workspace--history">
      <section className="card historyCard">
        <HistoryToolbar
          capability={capability}
          setCapability={(v) => {
            setCapability(v);
            setPage(1);
          }}
          qInput={qInput}
          setQInput={setQInput}
          onSearch={() => {
            setPage(1);
            setQ(qInput.trim());
          }}
          pageSize={pageSize}
          setPageSize={(n) => {
            setPage(1);
            setPageSize(n);
          }}
          total={total}
          page={page}
          pages={pages}
          busy={busy}
          onRefresh={() => void load(page)}
          onPrev={() => void load(page - 1)}
          onNext={() => void load(page + 1)}
        />

        {error && <div className="alert alert--err">Error: {error}</div>}

        <div className="list historyCard__list">
          {items.map((it) => (
            <HistoryRow
              key={it.id}
              item={it}
              active={it.id === selectedId}
              busy={busy}
              deleting={deletingId === it.id}
              onSelect={() => setSelectedId(it.id)}
              onDelete={() => onDeleteItem(it)}
            />
          ))}
          {items.length === 0 && !busy && (
            <div className="placeholder">
              <div className="placeholder__title">暂无记录</div>
              <div className="placeholder__sub">生成图片、视频或音频后，会自动保存到 MinIO 并出现在这里。</div>
            </div>
          )}
        </div>
      </section>

      <section className="card resultsCard historyPreviewCard">
        <div className="card__head">
          <h2 className="card__title">结果预览</h2>
          {selected ? (
            <div className="chips" style={{ marginLeft: "auto" }}>
              <a className="link" href={selected.url} target="_blank" rel="noreferrer">
                open
              </a>
              {selected.url.startsWith("/api/assets/") ? (
                <a className="btn btn--ghost" href={`${selected.url}?download=1`} title="下载资源">
                  下载
                </a>
              ) : null}
            </div>
          ) : (
            <span className="muted">-</span>
          )}
        </div>

        {selected ? (
          selected.capability === "video" ? (
            <div className="panel__media">
              <video className="video" controls src={selected.url} />
            </div>
          ) : selected.capability === "audio" ? (
            <div className="panel__media">
              <audio className="audioPlayer" controls src={selected.url} />
            </div>
          ) : (
            <a className="preview" href={selected.url} target="_blank" rel="noreferrer">
              <img className="preview__img" src={selected.url} alt={selected.prompt_preview || ""} />
            </a>
          )
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">选择一条记录预览</div>
            <div className="placeholder__sub">左侧列表选择图片、视频或音频。</div>
          </div>
        )}

        {selected?.error && <div className="alert alert--err">Error: {selected.error}</div>}
      </section>
    </div>
  );
}
