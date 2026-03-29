import React, { useEffect, useState } from "react";
import { api } from "../../api";
import { blobToDataURL } from "./fileBase64";

type HistoryImageItem = {
  id: number;
  url: string;
  prompt_preview?: string;
};

export function HistoryImageSelector(props: {
  selectedId: number | null;
  onPicked: (v: { id: number; url: string; dataUrl: string }) => void;
  onClear: () => void;
  hasValue: boolean;
  disabled?: boolean;
}) {
  const { selectedId, onPicked, onClear, hasValue, disabled } = props;
  const [items, setItems] = useState<HistoryImageItem[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [pickingId, setPickingId] = useState<number | null>(null);

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function load() {
    setBusy(true);
    setError(null);
    try {
      const res = await api.getHistory({ capability: "image", limit: 24, offset: 0 });
      setItems((res.items || []) as HistoryImageItem[]);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  async function pick(it: HistoryImageItem) {
    setPickingId(it.id);
    setError(null);
    try {
      const res = await fetch(it.url, { method: "GET" });
      if (!res.ok) throw new Error(`读取历史图片失败: HTTP ${res.status}`);
      const blob = await res.blob();
      if (!String(blob.type || "").startsWith("image/")) throw new Error("历史记录不是图片格式");
      const dataUrl = await blobToDataURL(blob);
      onPicked({ id: it.id, url: it.url, dataUrl });
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setPickingId(null);
    }
  }

  return (
    <div className="historypick">
      <div className="chips">
        <button className="btn btn--ghost" onClick={() => void load()} disabled={disabled || busy}>
          {busy ? "加载中..." : "刷新历史"}
        </button>
        <button className="btn btn--ghost" onClick={onClear} disabled={disabled || !hasValue}>
          清除
        </button>
        <div className="filepick__meta">{selectedId ? `已选择 #${selectedId}` : "请选择一张历史图片"}</div>
      </div>

      {error ? <div className="alert alert--err">Error: {error}</div> : null}

      <div className="historypick__grid">
        {items.map((it) => (
          <button
            key={it.id}
            className={it.id === selectedId ? "historypick__item historypick__item--active" : "historypick__item"}
            onClick={() => void pick(it)}
            title={it.prompt_preview || `#${it.id}`}
            disabled={disabled || pickingId === it.id}
          >
            <img className="historypick__img" src={it.url} alt={it.prompt_preview || `history-${it.id}`} />
            <span className="historypick__cap">#{it.id}</span>
          </button>
        ))}
        {!busy && items.length === 0 ? <div className="muted">暂无历史图片</div> : null}
      </div>
    </div>
  );
}
