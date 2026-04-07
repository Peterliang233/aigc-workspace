type StoredState = {
  images: any[];
  videos: any[];
  audios: any[];
};

const LS_KEY = "aigc_inflight_v1";

export function safeParseState(): StoredState {
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (!raw) return { images: [], videos: [], audios: [] };
    const obj = JSON.parse(raw);
    const images = Array.isArray(obj?.images) ? obj.images : [];
    const videos = Array.isArray(obj?.videos) ? obj.videos : [];
    const audios = Array.isArray(obj?.audios) ? obj.audios : [];
    return { images, videos, audios };
  } catch {
    return { images: [], videos: [], audios: [] };
  }
}

export function persistState(state: unknown) {
  try {
    localStorage.setItem(LS_KEY, JSON.stringify(state));
  } catch {
    return;
  }
}

export function makeId() {
  return typeof crypto !== "undefined" && "randomUUID" in crypto
    ? (crypto as any).randomUUID()
    : `${Date.now()}_${Math.random().toString(16).slice(2)}`;
}
