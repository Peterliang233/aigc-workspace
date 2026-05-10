import React, { useEffect, useMemo, useState } from "react";
import { api } from "../api";
import { ResultActions } from "./common/ResultActions";
import { type HistoryItem, loadHistoryItems } from "./history/historyItems";
import { StoryHistoryPreview } from "./history/StoryHistoryPreview";
import { HistoryRow } from "./history/HistoryRow";
import { HistoryToolbar } from "./history/HistoryToolbar";

export function HistoryStudio() {
  const [capability, setCapability] = useState<"all" | "image" | "video" | "audio" | "story">("all");
  const [q, setQ] = useState("");
  const [qInput, setQInput] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(5);
  const [total, setTotal] = useState(0);
  const [items, setItems] = useState<HistoryItem[]>([]);
  const [busy, setBusy] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const selected = useMemo(() => items.find((x) => x.id === selectedId) || null, [items, selectedId]);
  const pages = Math.max(1, Math.ceil(total / Math.max(1, pageSize)));
  async function load(targetPage = page) {
    setBusy(true);
    setError(null);
    try {
      const merged = await loadHistoryItems(capability, q);
      const start = (targetPage - 1) * pageSize;
      const list = merged.slice(start, start + pageSize);
      setItems(list);
      setTotal(merged.length);
      setPage(targetPage);
      const first = list[0];
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

  async function onDeleteItem(it: HistoryItem) {
    if (it.deletable === false) return;
    const ok = window.confirm(`确定永久删除记录 #${it.id} 吗？删除后不可恢复。`);
    if (!ok) return;
    setDeletingId(it.id);
    setError(null);
    try {
      await api.deleteHistory(Number(it.id));
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
              <div className="placeholder__sub">生成图片、视频、音频或故事后，会自动汇总到这里。</div>
            </div>
          )}
        </div>
      </section>

      <section className="card resultsCard historyPreviewCard">
        <div className="card__head">
          <h2 className="card__title">结果预览</h2>
          {selected ? (
            <div style={{ marginLeft: "auto" }}><ResultActions url={selected.url} /></div>
          ) : (
            <span className="muted">-</span>
          )}
        </div>

        {selected ? (
          <>
            <div className="panel">
              <div className="panel__row"><div className="k">Provider</div><div className="v">{selected.provider || "-"}</div></div>
              <div className="panel__row"><div className="k">Model</div><div className="v">{selected.model || "-"}</div></div>
              <div className="panel__row"><div className="k">Status</div><div className="v">{selected.status || "-"}</div></div>
            </div>
            {selected.capability === "story" && selected.story ? (
              <StoryHistoryPreview project={selected.story} />
            ) : selected.capability === "video" ? (
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
            )}
          </>
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">选择一条记录预览</div>
            <div className="placeholder__sub">左侧列表选择图片、视频、音频或故事。</div>
          </div>
        )}

        {selected?.error && <div className="alert alert--err">Error: {selected.error}</div>}
      </section>
    </div>
  );
}
