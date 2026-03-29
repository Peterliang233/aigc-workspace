import React, { useEffect, useRef, useState } from "react";
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
  const imageUrlRef = useRef("");
  const imageBase64Ref = useRef("");

  useEffect(() => {
    const s = splitValue(value);
    setImageUrl(s.imageUrl);
    setImageBase64(s.imageBase64);
    imageUrlRef.current = s.imageUrl;
    imageBase64Ref.current = s.imageBase64;
  }, [value]);

  function emitValue(nextUrl: string, nextBase64: string) {
    const u = String(nextUrl || "").trim();
    const b = String(nextBase64 || "");
    if (b.trim()) {
      onChange(b);
      return;
    }
    onChange(u);
  }

  return (
    <InitImagePicker
      imageUrl={imageUrl}
      onImageUrl={(v) => {
        imageUrlRef.current = String(v || "");
        setImageUrl(v);
        emitValue(imageUrlRef.current, imageBase64Ref.current);
      }}
      imageBase64={imageBase64}
      onImageBase64={(v) => {
        imageBase64Ref.current = String(v || "");
        setImageBase64(v);
        emitValue(imageUrlRef.current, imageBase64Ref.current);
      }}
      disabled={disabled}
      required={required}
    />
  );
}
