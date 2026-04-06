import React from "react";

type ActionLinkProps = {
  href: string;
  kind?: "open" | "download";
  compact?: boolean;
  download?: boolean | string;
  title?: string;
};

export function ActionLink(props: ActionLinkProps) {
  const { href, kind = "open", compact = false, download, title } = props;
  const isDownload = kind === "download";
  const cls = `actionBtn ${isDownload ? "actionBtn--secondary" : "actionBtn--primary"}${compact ? " actionBtn--sm" : ""}`;
  return (
    <a
      className={cls}
      href={href}
      target={isDownload ? undefined : "_blank"}
      rel={isDownload ? undefined : "noreferrer"}
      download={download}
      title={title || (isDownload ? "下载资源" : "查看资源")}
    >
      <span className="actionBtn__icon" aria-hidden="true">{isDownload ? "↓" : "↗"}</span>
      <span>{isDownload ? "下载" : "查看"}</span>
    </a>
  );
}

export function ResultActions(props: { url: string; downloadName?: string; compact?: boolean }) {
  const { url, downloadName, compact = false } = props;
  const isAsset = url.startsWith("/api/assets/");
  const downloadHref = isAsset ? `${url}?download=1` : url;
  const download = isAsset ? undefined : (downloadName || true);
  return (
    <div className={compact ? "actionBar actionBar--tight" : "actionBar"}>
      <ActionLink href={url} compact={compact} />
      <ActionLink href={downloadHref} kind="download" compact={compact} download={download} />
    </div>
  );
}
