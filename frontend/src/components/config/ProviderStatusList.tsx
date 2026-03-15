import React from "react";

export function ProviderStatusList(props: {
  providers: { id: string; label: string; configured: boolean; models: string[] }[];
}) {
  const { providers } = props;
  return (
    <section className="card resultsCard">
      <div className="card__head">
        <h2 className="card__title">当前状态</h2>
        <div className="badge">{providers.length} providers</div>
      </div>

      <div className="list">
        {providers.map((p) => (
          <div className="list__row" key={p.id}>
            <div>{p.label}</div>
            <div className={p.configured ? "pill pill--ok" : "pill"}>{p.configured ? "可用" : "未配置"}</div>
            <div className="muted">{(p.models || []).length} models</div>
          </div>
        ))}
      </div>

      <div className="placeholder" style={{ marginTop: 12 }}>
        <div className="placeholder__title">提示</div>
        <div className="placeholder__sub">如果你刚保存了配置，但下拉框没更新，回到图片生成页刷新一次即可。</div>
      </div>
    </section>
  );
}

