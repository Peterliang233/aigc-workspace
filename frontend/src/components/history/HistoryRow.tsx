import React from "react";
import { Icon } from "../../layout/Icon";

export type HistoryItem = {
  id: number;
  capability: "image" | "video" | "audio";
  provider: string;
  model?: string;
  prompt_preview?: string;
  bytes: number;
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

export function HistoryRow(props: {
  item: HistoryItem;
  active: boolean;
  busy?: boolean;
  deleting?: boolean;
  onSelect: () => void;
  onDelete: () => void;
}) {
  const { item, active, busy, deleting, onSelect, onDelete } = props;
  return (
    <div
      className={active ? "hrow hrow--active" : "hrow"}
      role="button"
      tabIndex={0}
      onClick={onSelect}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect();
        }
      }}
      title="Select"
    >
      <div className="hrow__top">
        <div className="mono">#{item.id}</div>
        <div className="pill">{item.capability}</div>
      </div>
      <div className="hrow__mid">
        <div className="hrow__meta" title={item.model ? `${item.provider} · ${item.model}` : item.provider}>
          <span className="muted">{item.provider}</span>
          {item.model ? <span className="hrow__sep">·</span> : null}
          {item.model ? <span className="hrow__model mono">{item.model}</span> : null}
        </div>
        <div className="muted">{fmtBytes(item.bytes)}</div>
      </div>
      <div className="hrow__bot">
        <div className="hrow__prompt">{item.prompt_preview || ""}</div>
        <div className="hrow__time mono">{item.created_at}</div>
      </div>
      <button
        className="hrow__del"
        title="永久删除"
        aria-label={`删除 #${item.id}`}
        disabled={!!busy || !!deleting}
        onClick={(e) => {
          e.stopPropagation();
          void onDelete();
        }}
      >
        {deleting ? "..." : <Icon name="trash" />}
      </button>
    </div>
  );
}
