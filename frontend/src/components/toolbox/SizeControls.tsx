import React from "react";
import { PRESETS, RATIOS } from "./presets";
import { clampInt } from "./target";

export const SIZE_KEY_RATIO = "__ratio__";

export function SizeControls(props: {
  sizeKey: string;
  onSizeKey: (v: string) => void;
  ratioKey: string;
  onRatioKey: (v: string) => void;
  longEdge: number;
  onLongEdge: (v: number) => void;
  disabled?: boolean;
}) {
  const { sizeKey, onSizeKey, ratioKey, onRatioKey, longEdge, onLongEdge, disabled } = props;
  return (
    <div style={{ display: "grid", gap: 12 }}>
      <label className="label">
        目标尺寸
        <select className="input" value={sizeKey} onChange={(e) => onSizeKey(e.target.value)} disabled={disabled}>
          {PRESETS.map((p) => (
            <option key={p.key} value={p.key}>
              {p.label}
            </option>
          ))}
          <option value={SIZE_KEY_RATIO}>自定义比例...</option>
        </select>
      </label>

      {sizeKey === SIZE_KEY_RATIO ? (
        <div className="row2">
          <label className="label">
            比例
            <select className="input" value={ratioKey} onChange={(e) => onRatioKey(e.target.value)} disabled={disabled}>
              {RATIOS.map((r) => (
                <option key={r.key} value={r.key}>
                  {r.label}
                </option>
              ))}
            </select>
          </label>
          <label className="label">
            长边(px)
            <input
              className="input"
              type="number"
              min={64}
              max={4096}
              step={1}
              value={longEdge}
              onChange={(e) => onLongEdge(clampInt(Number(e.target.value), 64, 4096))}
              disabled={disabled}
            />
          </label>
        </div>
      ) : null}
    </div>
  );
}
