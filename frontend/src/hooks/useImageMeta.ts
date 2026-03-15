import { useEffect, useMemo, useState } from "react";
import { api } from "../api";

type ProviderMeta = { id: string; label: string; configured: boolean; models: string[] };

export function useImageMeta() {
  const [metaLoading, setMetaLoading] = useState(false);
  const [providers, setProviders] = useState<ProviderMeta[]>([]);

  const [provider, setProvider] = useState<string>(() => localStorage.getItem("aigc_image_provider") || "");
  const [model, setModel] = useState<string>(() => localStorage.getItem("aigc_image_model") || "");
  const [customModel, setCustomModel] = useState<string>(() => localStorage.getItem("aigc_image_custom_model") || "");

  const providerInfo = useMemo(() => providers.find((p) => p.id === provider) || null, [providers, provider]);
  const modelList = providerInfo?.models || [];
  const useCustom = model === "__custom__" || modelList.length === 0;

  useEffect(() => {
    let mounted = true;
    async function load() {
      setMetaLoading(true);
      try {
        const res = await api.getImageMeta();
        if (!mounted) return;
        const list = (res.providers || []).slice().sort((a, b) => a.label.localeCompare(b.label));
        setProviders(list);

        const preferred = provider || res.default_provider || "mock";
        const hasPreferred = list.some((p) => p.id === preferred && p.configured);
        const fallback = list.find((p) => p.configured)?.id || "mock";
        const nextProvider = hasPreferred ? preferred : fallback;
        setProvider(nextProvider);

        const pi = list.find((p) => p.id === nextProvider);
        const models = pi?.models || [];
        if (!model) setModel(models[0] || "__custom__");
      } catch {
        if (mounted) {
          setProviders([{ id: "mock", label: "Mock(联调)", configured: true, models: [] }]);
          if (!provider) setProvider("mock");
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
    if (provider) localStorage.setItem("aigc_image_provider", provider);
  }, [provider]);
  useEffect(() => {
    if (model) localStorage.setItem("aigc_image_model", model);
  }, [model]);
  useEffect(() => {
    if (customModel) localStorage.setItem("aigc_image_custom_model", customModel);
  }, [customModel]);

  // When provider changes, reset model to first model (or custom) if current model isn't compatible.
  useEffect(() => {
    if (!providerInfo) return;
    const models = providerInfo.models || [];
    if (models.length === 0) {
      if (model !== "__custom__") setModel("__custom__");
      return;
    }
    if (model === "__custom__") return;
    if (!models.includes(model)) setModel(models[0]);
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
    modelList,
    useCustom
  };
}

