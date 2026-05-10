import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./styles.css";
import { GenerationProvider } from "./state/generation";
import { StoryVideoProvider } from "./state/storyvideo";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <GenerationProvider>
      <StoryVideoProvider>
        <App />
      </StoryVideoProvider>
    </GenerationProvider>
  </React.StrictMode>
);
