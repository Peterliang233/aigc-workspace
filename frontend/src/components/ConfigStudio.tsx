import React, { useEffect, useMemo, useState } from "react";
import { api, SettingsGetResponse } from "../api";

type ProvID = "siliconflow" | "wuyinkeji" | "openai_compatible";

export function ConfigStudio() {
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState<ProvID | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [okMsg, setOkMsg] = useState<string | null>(null);

  const [settings, setSettings] = useState<SettingsGetResponse | null>(null);
  const [meta, setMeta] = useState<
    { default_provider: string; providers: { id: string; label: string; configured: boolean; models: string[] }[] } | null
  >(null);

  // SiliconFlow editor state
  const [sfNewModel, setSfNewModel] = useState("");

  // 速创API editor state
  const [wyNewModel, setWyNewModel] = useState("");

  // OpenAI-compatible (optional)
  const [oaNewModel, setOaNewModel] = useState("");

  const provView = settings?.image_providers || {};

  const statusList = useMemo(() => {
    const list = meta?.providers || [];
    return list.slice().sort((a, b) => a.label.localeCompare(b.label));
  }, [meta]);

  async function reloadAll() {
    setLoading(true);
    setError(null);
    try {
      const [s, m] = await Promise.all([api.getSettings(), api.getImageMeta()]);
      setSettings(s);
      setMeta(m);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    reloadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function onAddModel(id: ProvID, model: string) {
    model = model.trim();
    if (!model) return;
    setSaving(id);
    setError(null);
    setOkMsg(null);
    try {
      await api.addImageModel(id, model);
      setOkMsg("已添加");
      await reloadAll();
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setSaving(null);
    }
  }

  async function onDeleteModel(id: ProvID, model: string) {
    setSaving(id);
    setError(null);
    setOkMsg(null);
    try {
      await api.deleteImageModel(id, model);
      setOkMsg("已删除");
      await reloadAll();
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setSaving(null);
    }
  }

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head">
          <h2 className="card__title">配置</h2>
        </div>

        <div className="form">
          <div className="alert">
            Base URL / API Key / 默认模型 建议通过部署环境配置；此页面用于管理模型列表（新增/删除）。
          </div>

          <div className="panel">
            <div className="panel__row">
              <div className="k">SiliconFlow</div>
              <div className="v">{provView["siliconflow"]?.api_key_set ? "已配置" : "未配置"}</div>
            </div>

            <div className="panel__row">
              <div className="k">Base URL</div>
              <div className="v mono">{provView["siliconflow"]?.base_url || "-"}</div>
            </div>
            <div className="panel__row">
              <div className="k">默认模型</div>
              <div className="v mono">{provView["siliconflow"]?.default_model || "-"}</div>
            </div>

            <div className="panel" style={{ marginTop: 10 }}>
              <div className="panel__row">
                <div className="k">当前模型</div>
                <div className="v">{(provView["siliconflow"]?.models || []).length}</div>
              </div>
              <div className="chips">
                {(provView["siliconflow"]?.models || []).map((m) => (
                  <button
                    key={m}
                    className="chip"
                    onClick={() => onDeleteModel("siliconflow", m)}
                    title="点击删除"
                    disabled={loading || saving !== null}
                  >
                    <span className="chip__text">{m}</span>
                    <span className="chip__x">del</span>
                  </button>
                ))}
                {(provView["siliconflow"]?.models || []).length === 0 && <div className="muted">-</div>}
              </div>
              <div className="row" style={{ marginTop: 10 }}>
                <label className="label" style={{ flex: 1 }}>
                  新增模型
                  <input
                    className="input"
                    value={sfNewModel}
                    onChange={(e) => setSfNewModel(e.target.value)}
                    placeholder="例如：Kwai-Kolors/Kolors"
                    disabled={loading}
                  />
                </label>
                <button
                  className="btn"
                  onClick={() => {
                    const v = sfNewModel.trim();
                    setSfNewModel("");
                    onAddModel("siliconflow", v);
                  }}
                  disabled={loading || saving !== null}
                >
                  添加
                </button>
              </div>
            </div>
          </div>

          <div className="panel">
            <div className="panel__row">
              <div className="k">速创API</div>
              <div className="v">{provView["wuyinkeji"]?.api_key_set ? "已配置" : "未配置"}</div>
            </div>

            <div className="panel__row">
              <div className="k">Base URL</div>
              <div className="v mono">{provView["wuyinkeji"]?.base_url || "-"}</div>
            </div>
            <div className="hint">
              速创API 通过 <span className="mono">/api/async/&lt;模型名&gt;</span> 动态路径调用，例如{" "}
              <span className="mono">image_nanoBanana_pro</span>。
            </div>

            <div className="panel" style={{ marginTop: 10 }}>
              <div className="panel__row">
                <div className="k">当前模型</div>
                <div className="v">{(provView["wuyinkeji"]?.models || []).length}</div>
              </div>
              <div className="chips">
                {(provView["wuyinkeji"]?.models || []).map((m) => (
                  <button
                    key={m}
                    className="chip"
                    onClick={() => onDeleteModel("wuyinkeji", m)}
                    title="点击删除"
                    disabled={loading || saving !== null}
                  >
                    <span className="chip__text">{m}</span>
                    <span className="chip__x">del</span>
                  </button>
                ))}
                {(provView["wuyinkeji"]?.models || []).length === 0 && <div className="muted">-</div>}
              </div>
              <div className="row" style={{ marginTop: 10 }}>
                <label className="label" style={{ flex: 1 }}>
                  新增模型
                  <input
                    className="input"
                    value={wyNewModel}
                    onChange={(e) => setWyNewModel(e.target.value)}
                    placeholder="例如：image_nanoBanana_pro"
                    disabled={loading}
                  />
                </label>
                <button
                  className="btn"
                  onClick={() => {
                    const v = wyNewModel.trim();
                    setWyNewModel("");
                    onAddModel("wuyinkeji", v);
                  }}
                  disabled={loading || saving !== null}
                >
                  添加
                </button>
              </div>
            </div>
          </div>

          <details className="panel">
            <summary className="summary">高级：OpenAI Compatible</summary>

            <div className="panel__row">
              <div className="k">状态</div>
              <div className="v">{provView["openai_compatible"]?.api_key_set ? "已配置" : "未配置"}</div>
            </div>

            <div className="panel__row">
              <div className="k">Base URL</div>
              <div className="v mono">{provView["openai_compatible"]?.base_url || "-"}</div>
            </div>
            <div className="panel__row">
              <div className="k">默认模型</div>
              <div className="v mono">{provView["openai_compatible"]?.default_model || "-"}</div>
            </div>

            <div className="panel" style={{ marginTop: 10 }}>
              <div className="panel__row">
                <div className="k">当前模型</div>
                <div className="v">{(provView["openai_compatible"]?.models || []).length}</div>
              </div>
              <div className="chips">
                {(provView["openai_compatible"]?.models || []).map((m) => (
                  <button
                    key={m}
                    className="chip"
                    onClick={() => onDeleteModel("openai_compatible", m)}
                    title="点击删除"
                    disabled={loading || saving !== null}
                  >
                    <span className="chip__text">{m}</span>
                    <span className="chip__x">del</span>
                  </button>
                ))}
                {(provView["openai_compatible"]?.models || []).length === 0 && <div className="muted">-</div>}
              </div>
              <div className="row" style={{ marginTop: 10 }}>
                <label className="label" style={{ flex: 1 }}>
                  新增模型
                  <input
                    className="input"
                    value={oaNewModel}
                    onChange={(e) => setOaNewModel(e.target.value)}
                    placeholder="例如：gpt-image-1"
                    disabled={loading}
                  />
                </label>
                <button
                  className="btn"
                  onClick={() => {
                    const v = oaNewModel.trim();
                    setOaNewModel("");
                    onAddModel("openai_compatible", v);
                  }}
                  disabled={loading || saving !== null}
                >
                  添加
                </button>
              </div>
            </div>
          </details>

          {error && <div className="alert alert--err">Error: {error}</div>}
          {okMsg && <div className="alert alert--ok">{okMsg}</div>}
        </div>
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">当前状态</h2>
          <div className="badge">{statusList.length} providers</div>
        </div>

        <div className="list">
          {statusList.map((p) => (
            <div className="list__row" key={p.id}>
              <div>{p.label}</div>
              <div className={p.configured ? "pill pill--ok" : "pill"}>{p.configured ? "可用" : "未配置"}</div>
              <div className="muted">{(p.models || []).length} models</div>
            </div>
          ))}
        </div>

        <div className="placeholder" style={{ marginTop: 12 }}>
          <div className="placeholder__title">提示</div>
          <div className="placeholder__sub">如果你刚保存了配置，但下拉框没更新，回到图片生成页刷新一次即可。</div>
        </div>
      </section>
    </div>
  );
}
