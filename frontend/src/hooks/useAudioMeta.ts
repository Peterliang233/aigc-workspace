import { useEffect, useMemo, useState } from "react";
import { api, ProviderModelMeta } from "../api";

type ProviderMeta = { id: string; label: string; configured: boolean; models: ProviderModelMeta[] };

export function useAudioMeta() {
  const [metaLoading, setMetaLoading] = useState(false);
  const [providers, setProviders] = useState<ProviderMeta[]>([]);
  const [provider, setProvider] = useState<string>(() => localStorage.getItem("aigc_audio_provider") || "");
  const [model, setModel] = useState<string>(() => localStorage.getItem("aigc_audio_model") || "");
  const [customModel, setCustomModel] = useState<string>(() => localStorage.getItem("aigc_audio_custom_model") || "");
  const providerInfo = useMemo(() => providers.find((p) => p.id === provider) || null, [providers, provider]);
  const models = providerInfo?.models || [];
  const modelList = models.map((m) => m.id);
  const useCustom = model === "__custom__" || modelList.length === 0;
  const selectedModelMeta = useMemo(() => models.find((m) => m.id === model) || null, [models, model]);

  useEffect(() => {
    let mounted = true;
    async function load() {
      setMetaLoading(true);
      try {
        const res = await api.getAudioMeta();
        if (!mounted) return;
        const list = (res.providers || []).slice().sort((a, b) => a.label.localeCompare(b.label));
        setProviders(list);
        const nextProvider = list.find((p) => p.id === (provider || res.default_provider) && p.configured)?.id || list.find((p) => p.configured)?.id || list[0]?.id || "bltcy";
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
    if (provider) localStorage.setItem("aigc_audio_provider", provider);
  }, [provider]);
  useEffect(() => {
    if (model) localStorage.setItem("aigc_audio_model", model);
  }, [model]);
  useEffect(() => {
    if (customModel) localStorage.setItem("aigc_audio_custom_model", customModel);
  }, [customModel]);
  useEffect(() => {
    if (!providerInfo) return;
    const mids = providerInfo.models.map((m) => m.id);
    if (mids.length === 0) {
      if (model !== "__custom__") setModel("__custom__");
      return;
    }
    if (model !== "__custom__" && !mids.includes(model)) setModel(mids[0]);
  }, [provider]);

  return { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, models, modelList, selectedModelMeta, useCustom };
}
