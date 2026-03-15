import React, { useEffect, useId, useMemo, useRef, useState } from "react";
import { Icon } from "../../layout/Icon";
import { fileToDataURL } from "./fileBase64";

export function InitImagePicker(props: {
  imageUrl: string;
  onImageUrl: (v: string) => void;
  imageBase64: string;
  onImageBase64: (v: string) => void;
  disabled?: boolean;
  required?: boolean;
}) {
  const { imageUrl, onImageUrl, imageBase64, onImageBase64, disabled, required } = props;
  const inputId = useId();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [fileName, setFileName] = useState("");
  const [filePreview, setFilePreview] = useState<string | null>(null);

  const active = useMemo(() => {
    if (imageBase64) return "upload";
    if (imageUrl.trim()) return "url";
    return "none";
  }, [imageBase64, imageUrl]);

  useEffect(() => {
    return () => {
      if (filePreview) URL.revokeObjectURL(filePreview);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function onPickFile(f: File | null) {
    if (!f) return;
    const preview = URL.createObjectURL(f);
    setFilePreview((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return preview;
    });
    setFileName(f.name);
    const { dataUrl } = await fileToDataURL(f);
    onImageBase64(dataUrl);
    onImageUrl("");
  }

  function clearUpload() {
    onImageBase64("");
    setFileName("");
    setFilePreview((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return null;
    });
    if (inputRef.current) inputRef.current.value = "";
  }

  return (
    <div className="label">
      <div>
        参考图片{" "}
        {required ? <span className="pill" style={{ marginLeft: 8 }}>必填</span> : <span className="muted">(可选)</span>}
      </div>

      <div className="filepick">
        <input
          className="filepick__input"
          id={inputId}
          ref={inputRef}
          type="file"
          accept="image/*"
          onChange={(e) => void onPickFile((e.target.files || [])[0] || null)}
          disabled={disabled}
        />
        <label className="btn btn--ghost filepick__btn" htmlFor={inputId} title="上传参考图片">
          <Icon name="upload" />
          <span>上传图片</span>
        </label>
        <button className="btn btn--ghost" onClick={clearUpload} disabled={disabled || !imageBase64} title="清除上传">
          清除
        </button>
        <div className="filepick__meta" title={fileName}>
          {imageBase64 ? `已上传：${fileName || "image"}` : "未上传"}
        </div>
      </div>

      <div className="row2" style={{ marginTop: 10 }}>
        <label className="label">
          或填写图片 URL
          <input
            className="input"
            value={imageUrl}
            onChange={(e) => {
              onImageUrl(e.target.value);
              if (e.target.value.trim()) onImageBase64("");
            }}
            placeholder="https://..."
            disabled={disabled}
          />
        </label>
        <div className="label">
          当前来源
          <input className="input" value={active === "none" ? "-" : active} readOnly />
        </div>
      </div>

      {filePreview ? (
        <div className="panel__media">
          <img className="preview__img" src={filePreview} alt="init" />
        </div>
      ) : null}
    </div>
  );
}
