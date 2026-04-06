import React, { useMemo, useState } from "react";
import { useAnimationMeta } from "../../hooks/useAnimationMeta";
import { useAnimationJobs } from "../../state/animation";
import { InitImageField } from "../video/InitImageField";

const FALLBACK_AR = ["16:9", "9:16", "1:1"];

export function AnimationJobForm(props: { latestStatus?: string }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, models, modelList, selectedModelMeta, useCustom } = useAnimationMeta();
  const { createAnimationJob } = useAnimationJobs();
  const [values, setValues] = useState<Record<string, string>>({ prompt: "一只机械狐狸在赛博城市霓虹街道中向前奔跑，镜头连续跟拍，氛围一致。", duration_seconds: "20", aspect_ratio: "16:9", seed: "", lead_image: "" });
  const aspectOptions = useMemo(() => {
    const field = (selectedModelMeta?.form?.fields || []).find((f) => String(f.key || "").trim() === "aspect_ratio");
    const opts = (field?.options || []).map((o) => o.value).filter(Boolean);
    return opts.length > 0 ? opts : FALLBACK_AR;
  }, [selectedModelMeta]);
  const missing = useMemo(() => {
    const out: string[] = [];
    if (!String(values.prompt || "").trim()) out.push("prompt");
    if (!String(values.duration_seconds || "").trim()) out.push("duration_seconds");
    if (useCustom && !customModel.trim()) out.push("model");
    return out;
  }, [values, useCustom, customModel]);

  async function onSubmit() {
    setBusy(true);
    setError(null);
    try {
      const seedNum = String(values.seed || "").trim() ? Number(String(values.seed).trim()) : undefined;
      await createAnimationJob({
        provider,
        model: useCustom ? customModel.trim() || undefined : model,
        prompt: String(values.prompt || ""),
        duration_seconds: Number(String(values.duration_seconds || "0")),
        aspect_ratio: String(values.aspect_ratio || "").trim() || undefined,
        lead_image: String(values.lead_image || "").trim() || undefined,
        seed: Number.isFinite(seedNum as any) ? (seedNum as number) : undefined
      });
    } catch (e: any) {
      setError(e?.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="card">
      <div className="card__head"><h2 className="card__title">动画工坊</h2></div>
      <div className="form">
        <div className="row2">
          <label className="label">平台<select className="input" value={provider} onChange={(e) => setProvider(e.target.value)} disabled={metaLoading}>{providers.map((p) => <option key={p.id} value={p.id} disabled={!p.configured}>{p.label}{!p.configured ? " (未配置)" : ""}</option>)}</select></label>
          <label className="label">模型{modelList.length > 0 ? <select className="input" value={model} onChange={(e) => setModel(e.target.value)}>{models.map((m) => <option key={m.id} value={m.id}>{m.label ? `${m.label} (${m.id})` : m.id}</option>)}<option value="__custom__">自定义...</option></select> : <input className="input" value={customModel} onChange={(e) => setCustomModel(e.target.value)} placeholder="输入 I2V 模型名称" />}</label>
        </div>
        {modelList.length > 0 && useCustom ? <label className="label">自定义模型<input className="input" value={customModel} onChange={(e) => setCustomModel(e.target.value)} placeholder="输入 I2V 模型名称" /></label> : null}
        <label className="label">动画描述<textarea className="textarea" rows={6} value={values.prompt || ""} onChange={(e) => setValues((prev) => ({ ...prev, prompt: e.target.value }))} placeholder="描述主体、场景、镜头语言和连续动作。" disabled={busy} /></label>
        <div className="row2">
          <label className="label">总时长（秒）<input className="input" type="number" min="1" max="180" value={values.duration_seconds || ""} onChange={(e) => setValues((prev) => ({ ...prev, duration_seconds: e.target.value }))} disabled={busy} /></label>
          <label className="label">画面比例<select className="input" value={values.aspect_ratio || aspectOptions[0]} onChange={(e) => setValues((prev) => ({ ...prev, aspect_ratio: e.target.value }))} disabled={busy}>{aspectOptions.map((v) => <option key={v} value={v}>{v}</option>)}</select></label>
        </div>
        <label className="label">Seed（可选）<input className="input" type="number" value={values.seed || ""} onChange={(e) => setValues((prev) => ({ ...prev, seed: e.target.value }))} placeholder="固定主体一致性" disabled={busy} /></label>
        <InitImageField value={values.lead_image || ""} onChange={(v) => setValues((prev) => ({ ...prev, lead_image: v }))} disabled={busy} />
        <button className="btn" disabled={busy || missing.length > 0} onClick={onSubmit}>{busy ? "Submitting..." : "Submit"}</button>
      </div>
      {error ? <div className="alert alert--err">Error: {error}</div> : null}
      {missing.length > 0 ? <div className="alert">缺少必填字段：{missing.join(", ")}</div> : null}
      {props.latestStatus && props.latestStatus !== "succeeded" && props.latestStatus !== "failed" ? <div className="alert">动画任务在后台持续执行，分段和拼接进度会自动刷新。</div> : null}
    </section>
  );
}
