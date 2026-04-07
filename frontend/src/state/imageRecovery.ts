import { api } from "../api";

type ImageTask = {
  id: string;
  status: "running" | "succeeded" | "failed";
  provider?: string;
  model?: string;
  prompt: string;
  image_url?: string;
  error?: string;
  created_at: number;
};

type HistoryItem = {
  id: number;
  provider: string;
  model?: string;
  prompt_preview?: string;
  url: string;
  created_at: string;
};

const STALE_MS = 15 * 60 * 1000;

export async function recoverRunningImages(tasks: ImageTask[]) {
  const running = tasks.filter((task) => task.status === "running");
  if (!running.length) return tasks;
  try {
    const res = await api.getHistory({
      capability: "image",
      page_size: Math.min(Math.max(running.length * 10, 10), 50)
    });
    return mergeHistory(tasks, res.items || []);
  } catch {
    return expireStale(tasks);
  }
}

function mergeHistory(tasks: ImageTask[], items: HistoryItem[]) {
  const used = new Set<number>();
  let changed = false;
  const next = tasks.map((task) => {
    if (task.status !== "running") return task;
    const match = findMatch(task, items, used);
    if (match) {
      used.add(match.id);
      changed = true;
      return { ...task, status: "succeeded" as const, image_url: match.url, error: undefined };
    }
    return task;
  });
  if (changed) return next;
  return expireStale(tasks);
}

function expireStale(tasks: ImageTask[]) {
  const now = Date.now();
  let changed = false;
  const next = tasks.map((task) => {
    if (task.status !== "running" || now - task.created_at < STALE_MS) return task;
    changed = true;
    return { ...task, status: "failed" as const, error: "前端已丢失该任务状态，请到历史记录确认结果或重新生成。" };
  });
  return changed ? next : tasks;
}

function findMatch(task: ImageTask, items: HistoryItem[], used: Set<number>) {
  const prompt = normalize(task.prompt);
  const provider = normalize(task.provider);
  const model = normalize(task.model);
  const earliest = task.created_at - 2 * 60 * 1000;
  for (const item of items) {
    if (used.has(item.id)) continue;
    if (provider && normalize(item.provider) !== provider) continue;
    if (model && normalize(item.model) && normalize(item.model) !== model) continue;
    if (!promptMatches(prompt, normalize(item.prompt_preview))) continue;
    const createdAt = Date.parse(item.created_at);
    if (Number.isFinite(createdAt) && createdAt < earliest) continue;
    return item;
  }
  return null;
}

function promptMatches(prompt: string, preview: string) {
  if (!prompt || !preview) return false;
  return prompt.startsWith(preview) || preview.startsWith(prompt);
}

function normalize(value?: string) {
  return String(value || "").trim().toLowerCase();
}
