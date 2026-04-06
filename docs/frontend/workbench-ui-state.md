# 前端工作台、布局与状态

## 主框架

应用入口：`frontend/src/App.tsx`

主框架分为两层：

- 左侧导航：`frontend/src/layout/Sidebar.tsx`
- 主内容区：按 tab 切换不同 Studio

当前 Studio 包括：

- `ImageStudio`
- `VideoStudio`
- `AudioStudio`
- `AnimationStudio`
- `ToolboxStudio`
- `HistoryStudio`

## 动态模型表单

前端不是写死每个平台的字段，而是由后端 `meta` 接口驱动：

- 图片：`frontend/src/hooks/useImageMeta.ts`
- 视频：`frontend/src/hooks/useVideoMeta.ts`
- 音频：`frontend/src/hooks/useAudioMeta.ts`
- 动画：`frontend/src/hooks/useAnimationMeta.ts`

字段渲染统一走：

- `frontend/src/components/form/ModelFields.tsx`
- `frontend/src/components/form/useApplyFieldDefaults.tsx`

因此“参数没显示出来”通常不是单个 Studio 的问题，要先检查 meta hook 和通用表单层。

## 全局生成状态

### 图片 / 视频 / 音频

状态入口：`frontend/src/state/generation.tsx`

职责：

- 把最近任务缓存到 `localStorage`
- 维护任务列表
- 轮询视频任务状态
- 为各个 Studio 提供统一的 `startImage` / `startAudio` / `createVideoJob`

### 动画

状态入口：`frontend/src/state/animation.tsx`

职责：

- 存动画任务列表
- 轮询动画任务状态
- 持久化到 `localStorage`

## 常见改动入口

### 改页面布局

先查：

- `frontend/src/App.tsx`
- `frontend/src/layout/Sidebar.tsx`
- `frontend/src/layout/MobileBar.tsx`
- `frontend/src/styles/*`

### 改视频生成表单

先查：

- `frontend/src/components/video/VideoJobForm.tsx`
- `frontend/src/components/video/videoRequest.ts`
- `frontend/src/components/video/videoFormDefaults.ts`

### 改历史列表展示

先查：

- `frontend/src/components/HistoryStudio.tsx`
- `frontend/src/components/history/HistoryRow.tsx`
- `frontend/src/components/history/HistoryToolbar.tsx`

### 改结果操作按钮

先查：

- `frontend/src/components/common/ResultActions.tsx`

## 常见问题定位

### 视频结果接口已经成功，前端还显示失败

先查：

- `frontend/src/state/generation.tsx` 的轮询逻辑
- `frontend/src/api.ts` 的返回字段映射
- `frontend/src/components/video/VideoResults.tsx` 的展示分支

### 表单字段被截断或值没传出去

先查：

- `ModelFields.tsx`
- `useApplyFieldDefaults.tsx`
- 对应 Studio 的 `onSubmit`
- 请求组装函数，例如 `videoRequest.ts`
