import React from "react";

export type IconName =
  | "image"
  | "video"
  | "audio"
  | "upload"
  | "toolbox"
  | "history"
  | "trash"
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
  if (name === "trash") {
    return (
      <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path d="M4.8 6h10.4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
        <path
          d="M7.2 6V4.9c0-.5.4-.9.9-.9h3.8c.5 0 .9.4.9.9V6"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        <path
          d="M6.4 6l.7 8.2c.05.55.5.96 1.05.96h3.7c.55 0 1-.41 1.05-.96L13.6 6"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path d="M8.8 8.8v4.5M11.2 8.8v4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
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
  if (name === "audio") {
    return (
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <path d="M5.4 11.6H3.8a1 1 0 0 1-1-1V9.4a1 1 0 0 1 1-1h1.6L9 5.6v8.8l-3.6-2.8Z" stroke="currentColor" strokeWidth="1.4" strokeLinejoin="round" />
        <path d="M12.4 7.2a4 4 0 0 1 0 5.6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
        <path d="M14.8 5.4a6.6 6.6 0 0 1 0 9.2" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
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
