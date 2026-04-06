import React, { useEffect, useId, useMemo, useRef, useState } from "react";
import { PRESETS, RATIOS, type Mode } from "./toolbox/presets";
import { outFileName, renderToPngBlob } from "./toolbox/render";
import { SizeControls, SIZE_KEY_RATIO } from "./toolbox/SizeControls";
import { targetFromRatio } from "./toolbox/target";
import { ResultActions } from "./common/ResultActions";
import { Icon } from "../layout/Icon";
export function ToolboxStudio() {
  const [srcFile, setSrcFile] = useState<File | null>(null);
  const [srcUrl, setSrcUrl] = useState<string | null>(null);
  const [outUrl, setOutUrl] = useState<string | null>(null);
  const [sizeKey, setSizeKey] = useState(PRESETS[0].key);
  const [ratioKey, setRatioKey] = useState(RATIOS[0].key);
  const [longEdge, setLongEdge] = useState(1024);
  const [mode, setMode] = useState<Mode>("cover");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [meta, setMeta] = useState<{ sw: number; sh: number } | null>(null);

  const fileInputId = useId();
  const workCanvasRef = useRef<HTMLCanvasElement | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const srcUrlRef = useRef<string | null>(null);
  const outUrlRef = useRef<string | null>(null);

  const preset = useMemo(() => PRESETS.find((p) => p.key === sizeKey) || PRESETS[0], [sizeKey]);
  const ratio = useMemo(() => RATIOS.find((r) => r.key === ratioKey) || RATIOS[0], [ratioKey]);
  const target = useMemo(
    () => (sizeKey === SIZE_KEY_RATIO ? targetFromRatio(ratio, longEdge) : { w: preset.w, h: preset.h }),
    [sizeKey, ratio, longEdge, preset.w, preset.h]
  );
  const outName = useMemo(() => outFileName(target), [target]);

  useEffect(() => {
    if (!srcFile) {
      setSrcUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
      return;
    }
    const next = URL.createObjectURL(srcFile);
    setSrcUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return next;
    });
    // srcFile change implies output should be recomputed
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [srcFile]);

  useEffect(() => {
    srcUrlRef.current = srcUrl;
  }, [srcUrl]);

  useEffect(() => {
    outUrlRef.current = outUrl;
  }, [outUrl]);

  useEffect(() => {
    return () => {
      if (srcUrlRef.current) URL.revokeObjectURL(srcUrlRef.current);
      if (outUrlRef.current) URL.revokeObjectURL(outUrlRef.current);
    };
  }, []);

  useEffect(() => {
    async function render() {
      if (!srcUrl) {
        setOutUrl((prev) => {
          if (prev) URL.revokeObjectURL(prev);
          return null;
        });
        setMeta(null);
        return;
      }
      setBusy(true);
      setError(null);
      try {
        const canvas = workCanvasRef.current || document.createElement("canvas");
        workCanvasRef.current = canvas;
        const { blob, sw, sh } = await renderToPngBlob({ srcUrl, size: target, mode, canvas });
        setMeta({ sw, sh });
        const nextOut = URL.createObjectURL(blob);
        setOutUrl((prev) => {
          if (prev) URL.revokeObjectURL(prev);
          return nextOut;
        });
      } catch (e: any) {
        setError(e?.message || String(e));
      } finally {
        setBusy(false);
      }
    }
    render();
  }, [srcUrl, target.w, target.h, mode]);

  function onReset() {
    setSrcFile(null);
    setError(null);
    setMeta(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
    setOutUrl((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return null;
    });
  }

  return (
    <div className="workspace">
      <section className="card">
        <div className="card__head">
          <h2 className="card__title">工具箱</h2>
        </div>
        <div className="form">
          <div className="label">
            <div>上传图片</div>
            <div className="filepick">
              <input
                className="filepick__input"
                id={fileInputId}
                type="file"
                accept="image/*"
                ref={fileInputRef}
                onChange={(e) => setSrcFile((e.target.files || [])[0] || null)}
              />
              <label className="btn btn--ghost filepick__btn" htmlFor={fileInputId} title="选择图片">
                <Icon name="upload" />
                <span>{srcFile ? "更换图片" : "选择图片"}</span>
              </label>
              <div className="filepick__meta" title={srcFile?.name || ""}>
                {srcFile ? srcFile.name : "支持 JPG/PNG/WebP 等常见格式"}
              </div>
            </div>
          </div>

          <div className="row2">
            <SizeControls
              sizeKey={sizeKey}
              onSizeKey={setSizeKey}
              ratioKey={ratioKey}
              onRatioKey={setRatioKey}
              longEdge={longEdge}
              onLongEdge={setLongEdge}
              disabled={busy}
            />
            <label className="label">
              适配方式
              <select className="input" value={mode} onChange={(e) => setMode(e.target.value as Mode)} disabled={busy}>
                <option value="cover">裁剪填充</option>
                <option value="contain">完整保留 (留白)</option>
              </select>
            </label>
          </div>

          <div className="row row--image">
            <label className="label">
              信息
              <input
                className="input"
                value={
                  meta ? `原图 ${meta.sw}x${meta.sh} -> 输出 ${target.w}x${target.h}` : `输出 ${target.w}x${target.h}`
                }
                readOnly
              />
            </label>
            <button className="btn btn--ghost" onClick={onReset} disabled={busy && !srcFile}>
              重置
            </button>
          </div>
        </div>

        {error && <div className="alert alert--err">Error: {error}</div>}
      </section>

      <section className="card resultsCard">
        <div className="card__head">
          <h2 className="card__title">预览</h2>
          {outUrl ? <div style={{ marginLeft: "auto" }}><ResultActions url={outUrl} downloadName={outName} /></div> : null}
        </div>

        {outUrl ? (
          <a className="preview" href={outUrl} target="_blank" rel="noreferrer" title="查看预览">
            <img className="preview__img" src={outUrl} alt="output" />
          </a>
        ) : (
          <div className="placeholder">
            <div className="placeholder__title">{busy ? "处理中..." : "请先上传一张图片"}</div>
            <div className="placeholder__sub">本工具不会保存历史，刷新页面即清空。</div>
          </div>
        )}
      </section>
    </div>
  );
}
