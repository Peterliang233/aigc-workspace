import { useEffect } from "react";
import type { ModelFormField } from "../../api";

export function useApplyFieldDefaults(
  fields: ModelFormField[],
  setValues: (fn: (prev: Record<string, string>) => Record<string, string>) => void,
  deps: any[]
) {
  useEffect(() => {
    if (!Array.isArray(fields) || fields.length === 0) return;
    setValues((prev) => {
      let changed = false;
      const next: Record<string, string> = { ...prev };
      for (const f of fields) {
        const k = String(f.key || "").trim();
        if (!k) continue;
        const cur = next[k] ?? "";
        const def = (f as any).default;
        if (!cur && def !== undefined && def !== null) {
          next[k] = String(def);
          changed = true;
        }
      }
      return changed ? next : prev;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);
}
