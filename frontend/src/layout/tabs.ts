export type Tab = "text" | "image" | "video" | "audio" | "animation-create" | "animation-records" | "toolbox" | "history";

export const TAB_LABEL: Record<Tab, string> = {
  text: "文本生成",
  image: "图片生成",
  video: "视频生成",
  audio: "音频生成",
  "animation-create": "故事工坊 · 新建项目",
  "animation-records": "故事工坊 · 项目记录",
  toolbox: "工具箱",
  history: "历史"
};
