import React, { useEffect, useMemo, useState } from "react";
import { api, SettingsGetResponse } from "../api";
import { ProviderModelsEditor } from "./config/ProviderModelsEditor";
import { ProviderStatusList } from "./config/ProviderStatusList";

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

  const [sfNewModel, setSfNewModel] = useState("");
  const [wyNewModel, setWyNewModel] = useState("");
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

          <ProviderModelsEditor
            title="SiliconFlow"
            apiKeySet={!!provView["siliconflow"]?.api_key_set}
            baseURL={provView["siliconflow"]?.base_url}
            defaultModel={provView["siliconflow"]?.default_model}
            models={provView["siliconflow"]?.models || []}
            newModel={sfNewModel}
            onNewModelChange={setSfNewModel}
            onAdd={() => {
              const v = sfNewModel.trim();
              setSfNewModel("");
              onAddModel("siliconflow", v);
            }}
            onDelete={(m) => onDeleteModel("siliconflow", m)}
            disabled={loading || saving !== null}
          />

          <ProviderModelsEditor
            title="速创API"
            apiKeySet={!!provView["wuyinkeji"]?.api_key_set}
            baseURL={provView["wuyinkeji"]?.base_url}
            models={provView["wuyinkeji"]?.models || []}
            hint={
              <>
                速创API 通过 <span className="mono">/api/async/&lt;模型名&gt;</span> 动态路径调用，例如{" "}
                <span className="mono">image_nanoBanana_pro</span>。
              </>
            }
            newModel={wyNewModel}
            onNewModelChange={setWyNewModel}
            onAdd={() => {
              const v = wyNewModel.trim();
              setWyNewModel("");
              onAddModel("wuyinkeji", v);
            }}
            onDelete={(m) => onDeleteModel("wuyinkeji", m)}
            disabled={loading || saving !== null}
          />

          <details className="panel">
            <summary className="summary">高级：OpenAI Compatible</summary>

            <ProviderModelsEditor
              title="OpenAI Compatible"
              apiKeySet={!!provView["openai_compatible"]?.api_key_set}
              baseURL={provView["openai_compatible"]?.base_url}
              defaultModel={provView["openai_compatible"]?.default_model}
              models={provView["openai_compatible"]?.models || []}
              newModel={oaNewModel}
              onNewModelChange={setOaNewModel}
              onAdd={() => {
                const v = oaNewModel.trim();
                setOaNewModel("");
                onAddModel("openai_compatible", v);
              }}
              onDelete={(m) => onDeleteModel("openai_compatible", m)}
              disabled={loading || saving !== null}
            />
          </details>

          {error && <div className="alert alert--err">Error: {error}</div>}
          {okMsg && <div className="alert alert--ok">{okMsg}</div>}
        </div>
      </section>

      <ProviderStatusList providers={statusList} />
    </div>
  );
}
