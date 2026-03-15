import React from "react";

export function ProviderModelsEditor(props: {
  title: string;
  apiKeySet: boolean;
  baseURL?: string;
  defaultModel?: string;
  hint?: React.ReactNode;
  models: string[];
  newModel: string;
  onNewModelChange: (v: string) => void;
  onAdd: () => void;
  onDelete: (m: string) => void;
  disabled: boolean;
}) {
  const { title, apiKeySet, baseURL, defaultModel, hint, models, newModel, onNewModelChange, onAdd, onDelete, disabled } =
    props;

  return (
    <div className="panel">
      <div className="panel__row">
        <div className="k">{title}</div>
        <div className="v">{apiKeySet ? "已配置" : "未配置"}</div>
      </div>

      <div className="panel__row">
        <div className="k">Base URL</div>
        <div className="v mono">{baseURL || "-"}</div>
      </div>

      {typeof defaultModel === "string" && (
        <div className="panel__row">
          <div className="k">默认模型</div>
          <div className="v mono">{defaultModel || "-"}</div>
        </div>
      )}

      {hint && <div className="hint">{hint}</div>}

      <div className="panel" style={{ marginTop: 10 }}>
        <div className="panel__row">
          <div className="k">当前模型</div>
          <div className="v">{models.length}</div>
        </div>
        <div className="chips">
          {models.map((m) => (
            <button key={m} className="chip" onClick={() => onDelete(m)} title="点击删除" disabled={disabled}>
              <span className="chip__text">{m}</span>
              <span className="chip__x">del</span>
            </button>
          ))}
          {models.length === 0 && <div className="muted">-</div>}
        </div>
        <div className="row" style={{ marginTop: 10 }}>
          <label className="label" style={{ flex: 1 }}>
            新增模型
            <input
              className="input"
              value={newModel}
              onChange={(e) => onNewModelChange(e.target.value)}
              placeholder="输入模型名称"
              disabled={disabled}
            />
          </label>
          <button className="btn" onClick={onAdd} disabled={disabled}>
            添加
          </button>
        </div>
      </div>
    </div>
  );
}

