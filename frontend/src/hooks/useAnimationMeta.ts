import { useEffect, useMemo, useState } from "react";
import { api, type ProviderModelMeta } from "../api";

type ProviderMeta = { id: string; label: string; configured: boolean; models: ProviderModelMeta[] };

function isContinuousModel(m: ProviderModelMeta) {
  return !!m.requires_image || !!m.form?.requires_image || !!(m.form?.fields || []).some((f) => String(f.key || "").trim() === "image" && !!f.required);
}

export function useAnimationMeta() {
  const [metaLoading, setMetaLoading] = useState(false);
  const [providers, setProviders] = useState<ProviderMeta[]>([]);
  const [provider, setProvider] = useState<string>(() => localStorage.getItem("aigc_animation_provider") || "");
  const [model, setModel] = useState<string>(() => localStorage.getItem("aigc_animation_model") || "");
  const [customModel, setCustomModel] = useState<string>(() => localStorage.getItem("aigc_animation_custom_model") || "");
  const providerInfo = useMemo(() => providers.find((p) => p.id === provider) || null, [providers, provider]);
  const models = useMemo(() => (providerInfo?.models || []).filter(isContinuousModel), [providerInfo]);
  const modelList = models.map((m) => m.id);
  const useCustom = model === "__custom__" || modelList.length === 0;
  const selectedModelMeta = useMemo(() => models.find((m) => m.id === model) || null, [models, model]);

  useEffect(() => {
    let mounted = true;
    async function load() {
      setMetaLoading(true);
      try {
        const res = await api.getVideoMeta();
        if (!mounted) return;
        const list = (res.providers || []).map((p) => ({ ...p, models: (p.models || []).filter(isContinuousModel) })).filter((p) => p.models.length > 0);
        setProviders(list);
        const nextProvider = list.find((p) => p.id === (provider || res.default_provider) && p.configured)?.id || list.find((p) => p.configured)?.id || list[0]?.id || "";
        setProvider(nextProvider);
        const mids = (list.find((p) => p.id === nextProvider)?.models || []).map((m) => m.id);
        if (!model) setModel(mids[0] || "__custom__");
      } finally {
        if (mounted) setMetaLoading(false);
      }
    }
    void load();
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    if (provider) localStorage.setItem("aigc_animation_provider", provider);
  }, [provider]);
  useEffect(() => {
    if (model) localStorage.setItem("aigc_animation_model", model);
  }, [model]);
  useEffect(() => {
    if (customModel) localStorage.setItem("aigc_animation_custom_model", customModel);
  }, [customModel]);
  useEffect(() => {
    if (!providerInfo) return;
    const mids = models.map((m) => m.id);
    if (mids.length === 0) {
      if (model !== "__custom__") setModel("__custom__");
      return;
    }
    if (model !== "__custom__" && !mids.includes(model)) setModel(mids[0]);
  }, [provider, providerInfo, model, models]);

  return { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, models, modelList, selectedModelMeta, useCustom };
}
