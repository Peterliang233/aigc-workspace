import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles.css";
import { AnimationProvider } from "./state/animation";
import { GenerationProvider } from "./state/generation";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <GenerationProvider>
      <AnimationProvider>
        <App />
      </AnimationProvider>
    </GenerationProvider>
  </React.StrictMode>
);
