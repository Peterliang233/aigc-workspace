import { useEffect, useMemo, useState } from "react";
import type { ProviderModelMeta } from "../api";
import { textApi } from "../api_text";

type ProviderMeta = { id: string; label: string; configured: boolean; models: ProviderModelMeta[] };

export function useTextMeta() {
  const [metaLoading, setMetaLoading] = useState(false);
  const [providers, setProviders] = useState<ProviderMeta[]>([]);
  const [provider, setProvider] = useState(() => localStorage.getItem("aigc_text_provider") || "");
  const [model, setModel] = useState(() => localStorage.getItem("aigc_text_model") || "");
  const [customModel, setCustomModel] = useState(() => localStorage.getItem("aigc_text_custom_model") || "");
  const providerInfo = useMemo(() => providers.find((item) => item.id === provider) || null, [provider, providers]);
  const models = providerInfo?.models || [];
  const modelList = models.map((item) => item.id);
  const useCustom = model === "__custom__" || modelList.length === 0;
  const selectedModelMeta = useMemo(() => models.find((item) => item.id === model) || null, [model, models]);

  useEffect(() => {
    let mounted = true;
    async function load() {
      setMetaLoading(true);
      try {
        const res = await textApi.getMeta();
        if (!mounted) return;
        const list = (res.providers || []).slice().sort((a, b) => a.label.localeCompare(b.label));
        setProviders(list);
        const nextProvider = list.find((item) => item.id === (provider || res.default_provider) && item.configured)?.id || list.find((item) => item.configured)?.id || list[0]?.id || "bltcy";
        setProvider(nextProvider);
        const mids = (list.find((item) => item.id === nextProvider)?.models || []).map((item) => item.id);
        if (!model) setModel(mids[0] || "__custom__");
      } finally {
        if (mounted) setMetaLoading(false);
      }
    }
    void load();
    return () => { mounted = false; };
  }, []);

  useEffect(() => { if (provider) localStorage.setItem("aigc_text_provider", provider); }, [provider]);
  useEffect(() => { if (model) localStorage.setItem("aigc_text_model", model); }, [model]);
  useEffect(() => { if (customModel) localStorage.setItem("aigc_text_custom_model", customModel); }, [customModel]);
  useEffect(() => {
    if (!providerInfo) return;
    const mids = providerInfo.models.map((item) => item.id);
    if (mids.length === 0) return setModel("__custom__");
    if (model !== "__custom__" && !mids.includes(model)) setModel(mids[0]);
  }, [model, provider, providerInfo]);

  return { metaLoading, providers, provider, setProvider, model, setModel, customModel, setCustomModel, models, modelList, selectedModelMeta, useCustom };
}
