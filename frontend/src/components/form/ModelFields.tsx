import React from "react";
import type { ModelFormField } from "../../api";

export function ModelFields(props: {
  fields: ModelFormField[];
  values: Record<string, string>;
  onChange: (key: string, value: string) => void;
  disabled?: boolean;
  custom?: (f: ModelFormField) => React.ReactNode | null;
}) {
  const { fields, values, onChange, disabled, custom } = props;

  return (
    <>
      {fields.map((f) => {
        const k = String(f.key || "").trim();
        if (!k) return null;

        const customNode = custom ? custom(f) : null;
        if (customNode) return <React.Fragment key={k}>{customNode}</React.Fragment>;

        const label = f.label || k;
        const val = values[k] ?? "";

        const head = (
          <>
            {label}
            {f.required ? <span className="pill" style={{ marginLeft: 8 }}>必填</span> : null}
          </>
        );

        if (f.type === "textarea") {
          return (
            <label className="label" key={k}>
              {head}
              <textarea
                className="textarea"
                value={val}
                onChange={(e) => onChange(k, e.target.value)}
                rows={typeof f.rows === "number" && f.rows > 0 ? f.rows : 4}
                placeholder={f.placeholder || ""}
                disabled={disabled}
              />
            </label>
          );
        }

        if (f.type === "select") {
          const opts = Array.isArray(f.options) ? f.options : [];
          return (
            <label className="label" key={k}>
              {head}
              <select className="input" value={val} onChange={(e) => onChange(k, e.target.value)} disabled={disabled}>
                {opts.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label || o.value}
                  </option>
                ))}
              </select>
            </label>
          );
        }

        const type = f.type === "number" ? "number" : "text";
        return (
          <label className="label" key={k}>
            {head}
            <input
              className="input"
              type={type}
              value={val}
              onChange={(e) => onChange(k, e.target.value)}
              placeholder={f.placeholder || ""}
              disabled={disabled}
            />
          </label>
        );
      })}
    </>
  );
}

