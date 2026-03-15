import { useEffect, useMemo, useState } from "react";
import { api } from "../api";

type ModelMeta = { id: string; label?: string; requires_image?: boolean };
type ProviderMeta = { id: string; label: string; configured: boolean; models: ModelMeta[] };

export function useVideoMeta() {
  const [metaLoading, setMetaLoading] = useState(false);
  const [providers, setProviders] = useState<ProviderMeta[]>([]);

  const [provider, setProvider] = useState<string>(() => localStorage.getItem("aigc_video_provider") || "");
  const [model, setModel] = useState<string>(() => localStorage.getItem("aigc_video_model") || "");
  const [customModel, setCustomModel] = useState<string>(() => localStorage.getItem("aigc_video_custom_model") || "");

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
        const res = await api.getVideoMeta();
        if (!mounted) return;
        const list = (res.providers || []).slice().sort((a, b) => a.label.localeCompare(b.label));
        setProviders(list);

        const preferred = provider || res.default_provider || "siliconflow";
        const hasPreferred = list.some((p) => p.id === preferred && p.configured);
        const fallback = list.find((p) => p.configured)?.id || preferred;
        const nextProvider = hasPreferred ? preferred : fallback;
        setProvider(nextProvider);

        const pi = list.find((p) => p.id === nextProvider);
        const mids = (pi?.models || []).map((m) => m.id);
        if (!model) setModel(mids[0] || "__custom__");
      } catch {
        if (mounted) {
          // keep a minimal fallback (SiliconFlow) so UI remains usable.
          setProviders([{ id: "siliconflow", label: "SiliconFlow", configured: false, models: [] }]);
          if (!provider) setProvider("siliconflow");
          if (!model) setModel("__custom__");
        }
      } finally {
        if (mounted) setMetaLoading(false);
      }
    }
    load();
    return () => {
      mounted = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (provider) localStorage.setItem("aigc_video_provider", provider);
  }, [provider]);
  useEffect(() => {
    if (model) localStorage.setItem("aigc_video_model", model);
  }, [model]);
  useEffect(() => {
    if (customModel) localStorage.setItem("aigc_video_custom_model", customModel);
  }, [customModel]);

  useEffect(() => {
    if (!providerInfo) return;
    const mids = (providerInfo.models || []).map((m) => m.id);
    if (mids.length === 0) {
      if (model !== "__custom__") setModel("__custom__");
      return;
    }
    if (model === "__custom__") return;
    if (!mids.includes(model)) setModel(mids[0]);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [provider]);

  return {
    metaLoading,
    providers,
    provider,
    setProvider,
    model,
    setModel,
    customModel,
    setCustomModel,
    models,
    modelList,
    selectedModelMeta,
    useCustom
  };
}
