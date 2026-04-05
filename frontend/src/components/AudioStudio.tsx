import React, { useMemo, useState } from "react";
import type { ModelFormField } from "../api";
import { useAudioMeta } from "../hooks/useAudioMeta";
import { useGeneration } from "../state/generation";
import { ModelFields } from "./form/ModelFields";
import { useApplyFieldDefaults } from "./form/useApplyFieldDefaults";

export function AudioStudio() {
  const { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, models, modelList, selectedModelMeta, useCustom } = useAudioMeta();
  const { state, startAudio } = useGeneration();
  const latest = useMemo(() => state.audios[0] || null, [state.audios]);
  const busy = state.audios.some((t) => t.status === "running");
  const fields: ModelFormField[] = useMemo(() => {
    const fs = selectedModelMeta?.form?.fields || [];
    if (fs.length > 0) return fs;
    return [
      { key: "input", label: "文本内容", type: "textarea", required: true, rows: 6 },
      { key: "voice", label: "音色", type: "text", required: true, placeholder: "alloy" }
    ];
  }, [selectedModelMeta]);
  const [values, setValues] = useState<Record<string, string>>(() => ({
    input: "你好，欢迎使用 AIGC Studio，这是一段新的语音示例。",
    voice: "alloy",
    response_format: "mp3",
    speed: "1"
  }));
  useApplyFieldDefaults(fields, setValues, [provider, model, customModel]);
  const missing = useMemo(() => fields.filter((f) => f.required).map((f) => String(f.key || "").trim()).filter((k) => k && !String(values[k] || "").trim()), [fields, values]);

  function onGenerate() {
    const pickedModel = useCustom ? customModel.trim() : model;
    const speedNum = String(values["speed"] || "").trim() ? Number(String(values["speed"]).trim()) : undefined;
    startAudio({
      provider,
      model: pickedModel || undefined,
      input: String(values["input"] || ""),
      voice: String(values["voice"] || "").trim() || undefined,
      response_format: String(values["response_format"] || "").trim() || undefined,
      speed: Number.isFinite(speedNum as any) ? (speedNum as number) : undefined
    });
  }

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head">
          <h2 className="card__title">音频生成</h2>
        </div>
        <div className="form">
          <div className="row2">
            <label className="label">
              平台
              <select className="input" value={provider} onChange={(e) => setProvider(e.target.value)} disabled={metaLoading}>
                {providers.map((p) => (
                  <option key={p.id} value={p.id} disabled={!p.configured}>
                    {p.label}
                    {!p.configured ? " (未配置)" : ""}
                  </option>
                ))}
              </select>
            </label>
            <label className="label">
              模型
              {modelList.length > 0 ? (
                <select className="input" value={model} onChange={(e) => setModel(e.target.value)}>
                  {models.map((m) => (
                    <option key={m.id} value={m.id}>{m.label ? `${m.label} (${m.id})` : m.id}</option>
                  ))}
                  <option value="__custom__">自定义...</option>
                </select>
              ) : (
                <input className="input" value={customModel} onChange={(e) => setCustomModel(e.target.value)} placeholder="输入模型名称" />
              )}
            </label>
          </div>
          {modelList.length > 0 && useCustom ? (
            <label className="label">
              自定义模型
              <input className="input" value={customModel} onChange={(e) => setCustomModel(e.target.value)} placeholder="输入模型名称" />
            </label>
          ) : null}
          <ModelFields fields={fields} values={values} onChange={(k, v) => setValues((prev) => ({ ...prev, [k]: v }))} disabled={busy} />
          <button className="btn" disabled={busy || missing.length > 0} onClick={onGenerate}>{busy ? "Generating..." : "Generate"}</button>
        </div>
        {latest?.status === "failed" ? <div className="alert alert--err">Error: {latest.error || "failed"}</div> : null}
        {missing.length > 0 ? <div className="alert">缺少必填字段：{missing.join(", ")}</div> : null}
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">生成结果</h2>
          {latest?.audio_url ? (
            <div className="chips" style={{ marginLeft: "auto" }}>
              <a className="link" href={latest.audio_url} target="_blank" rel="noreferrer">open</a>
              <a className="btn btn--ghost" href={latest.audio_url.startsWith("/api/assets/") ? `${latest.audio_url}?download=1` : latest.audio_url}>下载</a>
            </div>
          ) : null}
        </div>
        {latest?.audio_url ? (
          <div className="panel">
            <div className="panel__row"><div className="k">Model</div><div className="v">{latest.model || "-"}</div></div>
            <div className="panel__row"><div className="k">Voice</div><div className="v">{latest.voice || "-"}</div></div>
            <div className="panel__media"><audio className="audioPlayer" controls src={latest.audio_url} /></div>
          </div>
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">{busy ? "生成中..." : "结果会显示在这里"}</div>
            <div className="placeholder__sub">{busy ? "完成后会自动展示音频播放器。" : "点击 Generate 生成语音。"}</div>
          </div>
        )}
      </section>
    </div>
  );
}
