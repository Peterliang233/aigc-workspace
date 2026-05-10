import { api } from "../../api";
import { storyVideoApi, type StoryVideoProject } from "../../api_storyvideo";

export type HistoryCapability = "all" | "image" | "video" | "audio" | "story";

export type HistoryItem = {
  id: string;
  capability: "image" | "video" | "audio" | "story";
  provider: string;
  model?: string;
  status: string;
  error?: string;
  prompt_preview?: string;
  content_type: string;
  bytes: number;
  url: string;
  created_at: string;
  title?: string;
  deletable?: boolean;
  story?: StoryVideoProject;
};

function assetIdFromURL(url?: string) {
  const match = String(url || "").match(/\/api\/assets\/(\d+)/);
  return match?.[1] || "";
}

function storyAssetIDs(project: StoryVideoProject) {
  const ids = [assetIdFromURL(project.audio_url), assetIdFromURL(project.video_url)];
  for (const shot of project.shots || []) ids.push(assetIdFromURL(shot.image_url));
  return ids.filter(Boolean);
}

export async function loadHistoryItems(capability: HistoryCapability, q: string) {
  const [assetsRes, storiesRes] = await Promise.all([api.getHistory({ limit: 200 }), storyVideoApi.list()]);
  const stories = await Promise.all((storiesRes.items || []).map(async (item) => {
    try {
      return await storyVideoApi.get(item.id);
    } catch {
      return item;
    }
  }));
  const storyAssetIDsSet = new Set(stories.flatMap(storyAssetIDs));
  const assetItems: HistoryItem[] = (assetsRes.items || [])
    .filter((item) => !storyAssetIDsSet.has(String(item.id)))
    .map((item) => ({ ...item, id: String(item.id), deletable: true }));
  const storyItems: HistoryItem[] = stories.map((item) => ({
    id: item.id,
    capability: "story",
    provider: item.planner_provider || "storyvideo",
    model: item.planner_model,
    status: item.status,
    error: item.error,
    prompt_preview: item.summary || (item.keywords || []).join(" / "),
    content_type: "application/story+json",
    bytes: 0,
    url: item.video_url || item.audio_url || item.shots?.[0]?.image_url || "",
    created_at: item.created_at || "",
    title: item.title || item.id,
    deletable: false,
    story: item
  }));
  return [...assetItems, ...storyItems]
    .filter((item) => {
      if (capability !== "all" && item.capability !== capability) return false;
      if (!q) return true;
      const hay = [item.id, item.capability, item.provider, item.model, item.prompt_preview, item.title].join(" ").toLowerCase();
      return hay.includes(q.toLowerCase());
    })
    .sort((a, b) => String(b.created_at).localeCompare(String(a.created_at)));
}
