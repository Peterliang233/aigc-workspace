export async function fileToDataURL(file: File): Promise<{ dataUrl: string; mime: string }> {
  const mime = file.type || "application/octet-stream";
  const dataUrl = await blobToDataURL(file);
  return { dataUrl, mime };
}

export async function blobToDataURL(blob: Blob): Promise<string> {
  return await new Promise<string>((resolve, reject) => {
    const r = new FileReader();
    r.onload = () => resolve(String(r.result || ""));
    r.onerror = () => reject(new Error("读取文件失败"));
    r.readAsDataURL(blob);
  });
}
