import React, { useEffect, useId, useRef, useState } from "react";
import { Icon } from "../../layout/Icon";
import { fileToDataURL } from "./fileBase64";
import { HistoryImageSelector } from "./HistoryImageSelector";

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

  const [mode, setMode] = useState<"upload" | "history">("upload");
  const [fileName, setFileName] = useState("");
  const [localPreview, setLocalPreview] = useState<string | null>(null);
  const [historyPreview, setHistoryPreview] = useState<string | null>(null);
  const [pickedHistoryId, setPickedHistoryId] = useState<number | null>(null);

  useEffect(() => {
    return () => {
      if (localPreview) URL.revokeObjectURL(localPreview);
    };
  }, [localPreview]);

  const previewUrl = localPreview || historyPreview;
  const active = imageBase64 ? (pickedHistoryId ? "history" : "upload") : imageUrl.trim() ? "url" : "none";

  async function onPickFile(f: File | null) {
    if (!f) return;
    setMode("upload");
    const nextPreview = URL.createObjectURL(f);
    setLocalPreview((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return nextPreview;
    });
    setHistoryPreview(null);
    setPickedHistoryId(null);
    setFileName(f.name);
    const { dataUrl } = await fileToDataURL(f);
    onImageBase64(dataUrl);
    onImageUrl("");
  }

  function clearImage() {
    onImageBase64("");
    onImageUrl("");
    setFileName("");
    setPickedHistoryId(null);
    setHistoryPreview(null);
    setLocalPreview((prev) => {
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

      <div className="chips">
        <button
          className={mode === "upload" ? "chip" : "chip chip--ghost"}
          onClick={() => setMode("upload")}
          disabled={disabled}
        >
          <span className="chip__text">本地上传</span>
        </button>
        <button
          className={mode === "history" ? "chip" : "chip chip--ghost"}
          onClick={() => setMode("history")}
          disabled={disabled}
        >
          <span className="chip__text">历史图片</span>
        </button>
      </div>

      {mode === "upload" ? (
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
          <button className="btn btn--ghost" onClick={clearImage} disabled={disabled || !imageBase64} title="清除上传">
            清除
          </button>
          <div className="filepick__meta" title={fileName}>
            {imageBase64 ? `已上传：${fileName || "image"}` : "未上传"}
          </div>
        </div>
      ) : (
        <HistoryImageSelector
          selectedId={pickedHistoryId}
          hasValue={!!imageBase64}
          onClear={clearImage}
          disabled={disabled}
          onPicked={(v) => {
            setMode("history");
            setHistoryPreview(v.url);
            setPickedHistoryId(v.id);
            setFileName(`历史图片 #${v.id}`);
            setLocalPreview((prev) => {
              if (prev) URL.revokeObjectURL(prev);
              return null;
            });
            onImageBase64(v.dataUrl);
            onImageUrl("");
          }}
        />
      )}

      <div className="filepick__meta" title={fileName}>
        当前来源：{active === "none" ? "-" : active}
      </div>

      {previewUrl ? (
        <div className="panel__media">
          <img className="preview__img" src={previewUrl} alt="init" />
        </div>
      ) : null}
    </div>
  );
}
