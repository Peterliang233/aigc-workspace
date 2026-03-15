import React from "react";

export type IconName =
  | "image"
  | "video"
  | "settings"
  | "upload"
  | "toolbox"
  | "history"
  | "menu"
  | "collapse"
  | "expand";

export function Icon({ name }: { name: IconName }) {
  if (name === "menu") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path d="M3 5h14M3 10h14M3 15h14" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
      </svg>
    );
  }
  if (name === "collapse") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M12 4l-6 6 6 6"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }
  if (name === "expand") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M8 4l6 6-6 6"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }
  if (name === "settings") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M8.2 3.4h3.6l.4 2.1a6.3 6.3 0 0 1 1.4.8l2-.8 1.8 3.1-1.6 1.4c.1.3.1.7.1 1.1s0 .8-.1 1.1l1.6 1.4-1.8 3.1-2-.8c-.4.3-.9.6-1.4.8l-.4 2.1H8.2l-.4-2.1a6.3 6.3 0 0 1-1.4-.8l-2 .8-1.8-3.1 1.6-1.4A4.8 4.8 0 0 1 4.8 10c0-.4 0-.8.1-1.1L3.3 7.5 5.1 4.4l2 .8c.4-.3.9-.6 1.4-.8l.4-2.1Z"
          stroke="currentColor"
          strokeWidth="1.2"
          strokeLinejoin="round"
        />
        <path d="M10 12.6a2.6 2.6 0 1 0 0-5.2 2.6 2.6 0 0 0 0 5.2Z" stroke="currentColor" strokeWidth="1.2" />
      </svg>
    );
  }
  if (name === "history") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path d="M10 4.2a5.8 5.8 0 1 1-5.2 3.2" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
        <path
          d="M4.2 4.6v3.2h3.2"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M10 7v3.4l2.2 1.3"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }
  if (name === "toolbox") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M6.2 7.2V6.4c0-1.05.85-1.9 1.9-1.9h3.8c1.05 0 1.9.85 1.9 1.9v.8"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinecap="round"
        />
        <path
          d="M4.8 7.2h10.4c.66 0 1.2.54 1.2 1.2v6.8c0 .66-.54 1.2-1.2 1.2H4.8c-.66 0-1.2-.54-1.2-1.2V8.4c0-.66.54-1.2 1.2-1.2Z"
          stroke="currentColor"
          strokeWidth="1.4"
        />
        <path d="M8 11h4" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
      </svg>
    );
  }
  if (name === "upload") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M10 12.8V4.8"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M7.2 7.6 10 4.8l2.8 2.8"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M4.6 12.6v2.2c0 .72.58 1.3 1.3 1.3h8.2c.72 0 1.3-.58 1.3-1.3v-2.2"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }
  if (name === "image") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path
          d="M4.5 5.8c0-.72.58-1.3 1.3-1.3h8.4c.72 0 1.3.58 1.3 1.3v8.4c0 .72-.58 1.3-1.3 1.3H5.8c-.72 0-1.3-.58-1.3-1.3V5.8Z"
          stroke="currentColor"
          strokeWidth="1.4"
        />
        <path
          d="M6.3 12.4l2.2-2.2 2.2 2.2 1.3-1.3 2.0 2.0"
          stroke="currentColor"
          strokeWidth="1.4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path d="M7.6 8.2h.01" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round" />
      </svg>
    );
  }
  // video
  return (
    <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <path
        d="M4.8 6.2c0-.66.54-1.2 1.2-1.2h6.2c.66 0 1.2.54 1.2 1.2v7.6c0 .66-.54 1.2-1.2 1.2H6c-.66 0-1.2-.54-1.2-1.2V6.2Z"
        stroke="currentColor"
        strokeWidth="1.4"
      />
      <path
        d="M13.4 8l2.8-1.6v7.2L13.4 12V8Z"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
    </svg>
  );
}
