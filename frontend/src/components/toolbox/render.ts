import type { Mode } from "./presets";
import type { TargetSize } from "./target";

type RenderResult = { blob: Blob; sw: number; sh: number };

function loadImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error("图片加载失败"));
    img.src = url;
  });
}

function toPngBlob(canvas: HTMLCanvasElement): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((b) => (b ? resolve(b) : reject(new Error("导出失败 (toBlob)"))), "image/png");
  });
}

function drawContain(ctx: CanvasRenderingContext2D, img: HTMLImageElement, sw: number, sh: number, sz: TargetSize) {
  const s = Math.min(sz.w / sw, sz.h / sh);
  const dw = Math.round(sw * s);
  const dh = Math.round(sh * s);
  const dx = Math.floor((sz.w - dw) / 2);
  const dy = Math.floor((sz.h - dh) / 2);
  ctx.drawImage(img, 0, 0, sw, sh, dx, dy, dw, dh);
}

function drawCover(ctx: CanvasRenderingContext2D, img: HTMLImageElement, sw: number, sh: number, sz: TargetSize) {
  const dstAR = sz.w / sz.h;
  const srcAR = sw / sh;
  let sx = 0;
  let sy = 0;
  let sW = sw;
  let sH = sh;
  if (srcAR > dstAR) {
    sW = Math.round(sh * dstAR);
    sx = Math.floor((sw - sW) / 2);
  } else if (srcAR < dstAR) {
    sH = Math.round(sw / dstAR);
    sy = Math.floor((sh - sH) / 2);
  }
  ctx.drawImage(img, sx, sy, sW, sH, 0, 0, sz.w, sz.h);
}

export function outFileName(sz: TargetSize) {
  return `toolbox_${sz.w}x${sz.h}.png`;
}

export async function renderToPngBlob(args: {
  srcUrl: string;
  size: TargetSize;
  mode: Mode;
  canvas?: HTMLCanvasElement | null;
}): Promise<RenderResult> {
  const img = await loadImage(args.srcUrl);
  const sw = img.naturalWidth || img.width;
  const sh = img.naturalHeight || img.height;

  const canvas = args.canvas || document.createElement("canvas");
  canvas.width = args.size.w;
  canvas.height = args.size.h;
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("Canvas 不可用");

  ctx.clearRect(0, 0, args.size.w, args.size.h);
  ctx.imageSmoothingEnabled = true;
  // @ts-ignore: older TS DOM libs may not include imageSmoothingQuality
  ctx.imageSmoothingQuality = "high";

  if (args.mode === "contain") drawContain(ctx, img, sw, sh, args.size);
  else drawCover(ctx, img, sw, sh, args.size);

  const blob = await toPngBlob(canvas);
  return { blob, sw, sh };
}
