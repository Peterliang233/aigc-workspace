import React from "react";
import type { StoryVideoEvent } from "../../api_storyvideo";

export function StoryVideoEvents(props: { events: StoryVideoEvent[] }) {
  return (
    <section className="card">
      <div className="card__head">
        <h2 className="card__title">过程日志</h2>
      </div>
      <div className="list">
        {props.events.map((event) => (
          <div key={event.id} className="panel">
            <div className="panel__row"><strong>{event.stage}</strong><span className="k">{new Date(event.created_at).toLocaleString()}</span></div>
            <div>{event.message}</div>
            <div className="k">{event.type}</div>
          </div>
        ))}
        {props.events.length === 0 ? <div className="panel">暂无过程日志。</div> : null}
      </div>
    </section>
  );
}
