import React, { useEffect, useMemo, useState } from "react";
import type { ProviderMetaResponse } from "../../api";
import { useStoryVideo } from "../../state/storyvideo";

const CACHE_KEY = "aigc_story_video_form_v1";

type CachedForm = {
  keywords?: string;
  theme?: string;
  tone?: string;
  duration?: number;
  aspectRatio?: string;
  plannerProv?: string;
  plannerModel?: string;
  imageProv?: string;
  audioProv?: string;
  imageModel?: string;
  audioModel?: string;
  audioVoice?: string;
};

function models(meta: ProviderMetaResponse | null, provider: string) {
  return meta?.providers?.find((item) => item.id === provider)?.models || [];
}

function defaultPlannerModel(provider: string) {
  if (provider === "siliconflow") return "Qwen/Qwen3-32B";
  if (provider === "wuyinkeji") return "gpt-4.1";
  return "gpt-5.4";
}

function fieldOptions(meta: ProviderMetaResponse | null, provider: string, model: string, key: string) {
  return models(meta, provider).find((item) => item.id === model)?.form?.fields?.find((field) => field.key === key)?.options || [];
}

function readCache() {
  try {
    if (typeof window === "undefined") return {} as CachedForm;
    return JSON.parse(localStorage.getItem(CACHE_KEY) || "{}") as CachedForm;
  } catch {
    return {} as CachedForm;
  }
}

export function StoryVideoForm(props: { imageMeta: ProviderMetaResponse | null; audioMeta: ProviderMetaResponse | null; textMeta: ProviderMetaResponse | null; busy: boolean; error: string }) {
  const { createDraft } = useStoryVideo();
  const cached = useMemo(() => readCache(), []);
  const plannerProvider = props.textMeta?.default_provider || props.textMeta?.providers?.find((item) => item.id === "bltcy")?.id || props.textMeta?.providers?.[0]?.id || "bltcy";
  const imageProvider = props.imageMeta?.default_provider || props.imageMeta?.providers?.[0]?.id || "siliconflow";
  const audioProvider = props.audioMeta?.default_provider || props.audioMeta?.providers?.[0]?.id || "bltcy";
  const [keywords, setKeywords] = useState(cached.keywords || "赛博城市, 失忆侦探, 清晨追凶");
  const [theme, setTheme] = useState(cached.theme || "悬疑短篇");
  const [tone, setTone] = useState(cached.tone || "电影感、克制、带一点反转");
  const [duration, setDuration] = useState(cached.duration || 35);
  const [aspectRatio, setAspectRatio] = useState(cached.aspectRatio || "16:9");
  const [plannerProv, setPlannerProv] = useState(cached.plannerProv || plannerProvider);
  const [plannerModel, setPlannerModel] = useState(cached.plannerModel || defaultPlannerModel(plannerProvider));
  const [imageProv, setImageProv] = useState(cached.imageProv || imageProvider);
  const [audioProv, setAudioProv] = useState(cached.audioProv || audioProvider);
  const plannerModels = useMemo(() => models(props.textMeta, plannerProv), [props.textMeta, plannerProv]);
  const imageModels = useMemo(() => models(props.imageMeta, imageProv), [props.imageMeta, imageProv]);
  const audioModels = useMemo(() => models(props.audioMeta, audioProv), [props.audioMeta, audioProv]);
  const [imageModel, setImageModel] = useState(cached.imageModel || "");
  const [audioModel, setAudioModel] = useState(cached.audioModel || "");
  const audioVoiceOptions = useMemo(() => fieldOptions(props.audioMeta, audioProv, audioModel, "voice"), [props.audioMeta, audioProv, audioModel]);
  const [audioVoice, setAudioVoice] = useState(cached.audioVoice || "");

  useEffect(() => {
    if (!props.textMeta?.providers?.some((item) => item.id === plannerProv)) setPlannerProv(plannerProvider);
  }, [plannerProv, plannerProvider, props.textMeta]);

  useEffect(() => {
    if (plannerModels.length > 0 && !plannerModels.some((item) => item.id === plannerModel)) setPlannerModel(plannerModels[0].id);
    if (plannerModels.length === 0 && !plannerModel.trim()) setPlannerModel(defaultPlannerModel(plannerProv));
  }, [plannerModel, plannerModels, plannerProv]);

  useEffect(() => {
    if (!props.imageMeta?.providers?.some((item) => item.id === imageProv)) setImageProv(imageProvider);
  }, [imageProv, imageProvider, props.imageMeta]);

  useEffect(() => {
    if (!props.audioMeta?.providers?.some((item) => item.id === audioProv)) setAudioProv(audioProvider);
  }, [audioProv, audioProvider, props.audioMeta]);

  useEffect(() => {
    if (!imageModels.some((item) => item.id === imageModel)) setImageModel(imageModels[0]?.id || "");
  }, [imageModel, imageModels]);

  useEffect(() => {
    if (!audioModels.some((item) => item.id === audioModel)) setAudioModel(audioModels[0]?.id || "");
  }, [audioModel, audioModels]);

  useEffect(() => {
    if (audioVoiceOptions.length > 0 && !audioVoiceOptions.some((item) => item.value === audioVoice)) setAudioVoice(audioVoiceOptions[0].value);
  }, [audioVoice, audioVoiceOptions]);

  useEffect(() => {
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      keywords, theme, tone, duration, aspectRatio,
      plannerProv, plannerModel, imageProv, audioProv, imageModel, audioModel, audioVoice
    }));
  }, [keywords, theme, tone, duration, aspectRatio, plannerProv, plannerModel, imageProv, audioProv, imageModel, audioModel, audioVoice]);

  async function onSubmit() {
    await createDraft({
      keywords: keywords.split(/[\n,，]/).map((item) => item.trim()).filter(Boolean),
      theme,
      tone,
      duration_seconds: duration,
      aspect_ratio: aspectRatio,
      planner_provider: plannerProv,
      planner_model: plannerModel || plannerModels[0]?.id || defaultPlannerModel(plannerProv),
      image_provider: imageProv,
      image_model: imageModel || imageModels[0]?.id,
      audio_provider: audioProv,
      audio_model: audioModel || audioModels[0]?.id,
      audio_voice: audioVoice || audioVoiceOptions[0]?.value
    });
  }

  return (
    <section className="card storyvideoForm">
      <div className="card__head">
        <h2 className="card__title">故事视频工坊</h2>
        <span className="badge">关键词 → 台本 → 图片 → 音频 → 视频</span>
      </div>
      <div className="form">
        {props.error ? <div className="alert alert--err">{props.error}</div> : null}
        <label className="label">关键词<textarea className="textarea" rows={4} value={keywords} onChange={(e) => setKeywords(e.target.value)} /></label>
        <div className="row2">
          <label className="label">主题<input className="input" value={theme} onChange={(e) => setTheme(e.target.value)} /></label>
          <label className="label">时长(秒)<input className="input" type="number" min={10} max={180} value={duration} onChange={(e) => setDuration(Number(e.target.value) || 30)} /></label>
        </div>
        <label className="label">风格<textarea className="textarea" rows={3} value={tone} onChange={(e) => setTone(e.target.value)} /></label>
        <label className="label">画幅<select className="input" value={aspectRatio} onChange={(e) => setAspectRatio(e.target.value)}><option>16:9</option><option>9:16</option><option>1:1</option><option>4:3</option><option>3:4</option></select></label>
        <div className="row2">
          <label className="label">台本平台<select className="input" value={plannerProv} onChange={(e) => setPlannerProv(e.target.value)}>{props.textMeta?.providers?.map((item) => <option key={item.id} value={item.id}>{item.label}</option>)}</select></label>
          <label className="label">台本模型<select className="input" value={plannerModel} onChange={(e) => setPlannerModel(e.target.value)}>{plannerModels.length > 0 ? plannerModels.map((item) => <option key={item.id} value={item.id}>{item.label ? `${item.label} (${item.id})` : item.id}</option>) : <option value={plannerModel || defaultPlannerModel(plannerProv)}>{plannerModel || defaultPlannerModel(plannerProv)}</option>}</select></label>
        </div>
        <div className="row2">
          <label className="label">图片平台<select className="input" value={imageProv} onChange={(e) => setImageProv(e.target.value)}>{props.imageMeta?.providers?.map((item) => <option key={item.id} value={item.id}>{item.label}</option>)}</select></label>
          <label className="label">图片模型<select className="input" value={imageModel} onChange={(e) => setImageModel(e.target.value)}>{imageModels.map((item) => <option key={item.id} value={item.id}>{item.label ? `${item.label} (${item.id})` : item.id}</option>)}</select></label>
        </div>
        <div className={audioVoiceOptions.length > 0 ? "storyvideoAudioRow" : "row2"}>
          <label className="label">音频平台<select className="input" value={audioProv} onChange={(e) => setAudioProv(e.target.value)}>{props.audioMeta?.providers?.map((item) => <option key={item.id} value={item.id}>{item.label}</option>)}</select></label>
          <label className="label">音频模型<select className="input" value={audioModel} onChange={(e) => setAudioModel(e.target.value)}>{audioModels.map((item) => <option key={item.id} value={item.id}>{item.label ? `${item.label} (${item.id})` : item.id}</option>)}</select></label>
          {audioVoiceOptions.length > 0 ? <label className="label">音色<select className="input" value={audioVoice} onChange={(e) => setAudioVoice(e.target.value)}>{audioVoiceOptions.map((item) => <option key={item.value} value={item.value}>{item.label ? `${item.label} (${item.value})` : item.value}</option>)}</select></label> : null}
        </div>
        <button className="btn" disabled={props.busy} onClick={onSubmit}>{props.busy ? "处理中..." : "生成草稿"}</button>
      </div>
    </section>
  );
}
