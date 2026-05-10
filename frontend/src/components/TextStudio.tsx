import React, { useMemo, useState } from "react";
import type { ModelFormField } from "../api";
import { textApi } from "../api_text";
import { useTextMeta } from "../hooks/useTextMeta";
import { ModelFields } from "./form/ModelFields";
import { useApplyFieldDefaults } from "./form/useApplyFieldDefaults";

type Result = { text: string; provider: string; model?: string } | null;

export function TextStudio() {
  const { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, models, modelList, selectedModelMeta, useCustom } = useTextMeta();
  const [values, setValues] = useState<Record<string, string>>({ prompt: "请帮我写一篇关于 AI 提效办公的中文短文。", system_prompt: "" });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<Result>(null);
  const fields: ModelFormField[] = useMemo(() => {
    const fs = selectedModelMeta?.form?.fields || [];
    if (fs.length > 0) return fs;
    return [
      { key: "prompt", label: "文本需求", type: "textarea", required: true, rows: 8 },
      { key: "system_prompt", label: "系统提示词", type: "textarea", rows: 4 },
      { key: "temperature", label: "温度", type: "number", default: 0.7 },
      { key: "max_tokens", label: "最大输出", type: "number", default: 1200 }
    ];
  }, [selectedModelMeta]);
  useApplyFieldDefaults(fields, setValues, [provider, model, customModel]);
  const missing = useMemo(() => fields.filter((item) => item.required).map((item) => String(item.key || "").trim()).filter((key) => key && !String(values[key] || "").trim()), [fields, values]);

  async function onGenerate() {
    setBusy(true);
    setError("");
    try {
      const pickedModel = useCustom ? customModel.trim() : model;
      const temperature = String(values.temperature || "").trim() ? Number(values.temperature) : undefined;
      const maxTokens = String(values.max_tokens || "").trim() ? Number(values.max_tokens) : undefined;
      const res = await textApi.generate({
        provider,
        model: pickedModel || undefined,
        prompt: String(values.prompt || ""),
        system_prompt: String(values.system_prompt || "").trim() || undefined,
        temperature: Number.isFinite(temperature as any) ? temperature : undefined,
        max_tokens: Number.isFinite(maxTokens as any) ? maxTokens : undefined
      });
      setResult(res);
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head"><h2 className="card__title">文本生成</h2></div>
        <div className="form">
          <div className="row2">
            <label className="label">平台<select className="input" value={provider} onChange={(e) => setProvider(e.target.value)} disabled={metaLoading}>{providers.map((item) => <option key={item.id} value={item.id}>{item.label}{!item.configured ? " (未配置)" : ""}</option>)}</select></label>
            <label className="label">模型{modelList.length > 0 ? <select className="input" value={model} onChange={(e) => setModel(e.target.value)}>{models.map((item) => <option key={item.id} value={item.id}>{item.label ? `${item.label} (${item.id})` : item.id}</option>)}<option value="__custom__">自定义...</option></select> : <input className="input" value={customModel} onChange={(e) => setCustomModel(e.target.value)} placeholder="输入模型名称" />}</label>
          </div>
          {modelList.length > 0 && useCustom ? <label className="label">自定义模型<input className="input" value={customModel} onChange={(e) => setCustomModel(e.target.value)} placeholder="输入模型名称" /></label> : null}
          <ModelFields fields={fields} values={values} onChange={(key, value) => setValues((prev) => ({ ...prev, [key]: value }))} disabled={busy} />
          <button className="btn" disabled={busy || missing.length > 0} onClick={onGenerate}>{busy ? "Generating..." : "Generate"}</button>
        </div>
        {error ? <div className="alert alert--err">Error: {error}</div> : null}
        {missing.length > 0 ? <div className="alert">缺少必填字段：{missing.join(", ")}</div> : null}
      </section>
      <section className="card resultsCard">
        <div className="card__head"><h2 className="card__title">生成结果</h2></div>
        {result?.text ? <div className="panel"><div className="panel__row"><div className="k">Model</div><div className="v">{result.model || "-"}</div></div><textarea className="textarea" rows={18} value={result.text} readOnly /></div> : <div className="placeholder"><div className="placeholder__title">{busy ? "生成中..." : "结果会显示在这里"}</div><div className="placeholder__sub">{busy ? "完成后会自动显示文本结果。" : "输入需求后点击 Generate。"}</div></div>}
      </section>
    </div>
  );
}
