import React, { useEffect, useState } from "react";
import { InitImagePicker } from "./InitImagePicker";

function splitValue(v: string): { imageUrl: string; imageBase64: string } {
  v = String(v || "").trim();
  if (!v) return { imageUrl: "", imageBase64: "" };
  if (v.startsWith("data:")) return { imageUrl: "", imageBase64: v };
  return { imageUrl: v, imageBase64: "" };
}

export function InitImageField(props: {
  value: string;
  onChange: (v: string) => void;
  disabled?: boolean;
  required?: boolean;
}) {
  const { value, onChange, disabled, required } = props;
  const [imageUrl, setImageUrl] = useState("");
  const [imageBase64, setImageBase64] = useState("");

  useEffect(() => {
    const s = splitValue(value);
    setImageUrl(s.imageUrl);
    setImageBase64(s.imageBase64);
  }, [value]);

  return (
    <InitImagePicker
      imageUrl={imageUrl}
      onImageUrl={(v) => {
        setImageUrl(v);
        if (String(v).trim()) onChange(String(v).trim());
        else if (!imageBase64) onChange("");
      }}
      imageBase64={imageBase64}
      onImageBase64={(v) => {
        setImageBase64(v);
        if (String(v).trim()) onChange(String(v));
        else if (!imageUrl.trim()) onChange("");
      }}
      disabled={disabled}
      required={required}
    />
  );
}

