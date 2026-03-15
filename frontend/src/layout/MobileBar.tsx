import React from "react";
import { Icon } from "./Icon";
import type { Tab } from "./tabs";
import { TAB_LABEL } from "./tabs";

export function MobileBar(props: { tab: Tab; onOpenMenu: () => void }) {
  const { tab, onOpenMenu } = props;
  return (
    <div className="mobilebar">
      <button className="mobilebar__btn" onClick={onOpenMenu} title="Open menu">
        <Icon name="menu" />
      </button>
      <div className="mobilebar__title">{TAB_LABEL[tab]}</div>
    </div>
  );
}

