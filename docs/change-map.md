# 需求到模块的快速索引

这个文件是任务入口索引。后续要改需求时，先按下面的“需求类型”定位模块，再进入对应文档和代码。

## 1. 页面与交互

| 需求 | 文档 | 核心代码入口 |
| --- | --- | --- |
| 左侧菜单、顶部栏、整体布局 | `docs/frontend/workbench-ui-state.md` | `frontend/src/App.tsx`, `frontend/src/layout/Sidebar.tsx`, `frontend/src/layout/MobileBar.tsx` |
| 某个 Studio 页面 UI | `docs/frontend/workbench-ui-state.md` | `frontend/src/components/ImageStudio.tsx`, `frontend/src/components/VideoStudio.tsx`, `frontend/src/components/AudioStudio.tsx`, `frontend/src/components/AnimationStudio.tsx`, `frontend/src/components/HistoryStudio.tsx` |
| 通用表单字段、动态模型表单 | `docs/frontend/workbench-ui-state.md` | `frontend/src/components/form/ModelFields.tsx`, `frontend/src/components/form/useApplyFieldDefaults.tsx` |
| 结果区按钮、查看/下载/open UI | `docs/frontend/workbench-ui-state.md` | `frontend/src/components/common/ResultActions.tsx`, `frontend/src/components/video/VideoResults.tsx`, `frontend/src/components/history/HistoryRow.tsx` |

## 2. 图片 / 视频 / 音频生成

| 需求 | 文档 | 核心代码入口 |
| --- | --- | --- |
| 图片生成接口 | `docs/backend/http-api-runtime.md` | `backend/internal/httpapi/images_generate.go` |
| 视频任务创建与查询 | `docs/backend/http-api-runtime.md` | `backend/internal/httpapi/videos_create.go`, `backend/internal/httpapi/videos_get.go` |
| 音频生成接口 | `docs/backend/http-api-runtime.md` | `backend/internal/httpapi/audios_generate.go` |
| Provider 运行时选择 | `docs/backend/providers-models.md` | `backend/internal/httpapi/providers_build.go`, `backend/internal/httpapi/providers_*_runtime.go` |
| 模型表单透出 | `docs/backend/providers-models.md` | `backend/models.json`, `backend/internal/httpapi/meta_*.go` |

## 3. 平台接入与模型扩展

| 需求 | 文档 | 核心代码入口 |
| --- | --- | --- |
| 新增 Provider | `docs/backend/providers-models.md` | `backend/internal/providers/<provider>/`, `backend/internal/httpapi/providers_build.go` |
| 新增某个平台的模型 | `docs/backend/providers-models.md` | `backend/models.json` |
| 表单必填校验/默认值 | `docs/backend/providers-models.md` | `backend/internal/httpapi/model_form_*.go` |
| 下游请求日志与返回体排查 | `docs/backend/http-api-runtime.md` | `backend/internal/logging/downstream.go`, 对应 provider 实现 |

## 4. 历史记录与资产存储

| 需求 | 文档 | 核心代码入口 |
| --- | --- | --- |
| 生成结果入库 | `docs/backend/assets-history.md` | `backend/internal/httpapi/images_generate.go`, `backend/internal/httpapi/videos_get.go`, `backend/internal/httpapi/audios_generate.go`, `backend/internal/httpapi/animations_media.go` |
| 远程 URL 下载并写 MinIO | `docs/backend/assets-history.md` | `backend/internal/assets/service_remote.go`, `backend/internal/blobstore/*` |
| 历史列表/删除/详情 | `docs/backend/assets-history.md` | `backend/internal/httpapi/history.go`, `backend/internal/httpapi/history_delete.go`, `backend/internal/httpapi/assets_stream.go` |

## 5. 动画工坊

| 需求 | 文档 | 核心代码入口 |
| --- | --- | --- |
| 动画任务创建 | `docs/features/animation-workshop.md` | `backend/internal/httpapi/animations_create.go` |
| Prompt 拆分/分镜规划 | `docs/features/animation-workshop.md` | `backend/internal/httpapi/animations_planner.go`, `backend/internal/animation/planner_client.go`, `backend/internal/animation/plan.go` |
| 分段视频生成与轮询 | `docs/features/animation-workshop.md` | `backend/internal/httpapi/animations_run.go` |
| 尾帧抽取、拼接、媒体工作器 | `docs/features/animation-workshop.md` | `backend/internal/animation/media_client*.go`, `backend/internal/httpapi/animations_media.go` |
| 前端动画任务状态展示 | `docs/features/animation-workshop.md` | `frontend/src/components/animation/*`, `frontend/src/state/animation.tsx` |

## 6. 系统级配置

| 需求 | 文档 | 核心代码入口 |
| --- | --- | --- |
| 服务启动、环境变量、总路由 | `docs/architecture/system-overview.md` | `backend/main.go`, `backend/internal/config/config.go`, `backend/internal/httpapi/router.go` |
| 前端应用挂载与主框架 | `docs/frontend/workbench-ui-state.md` | `frontend/src/App.tsx` |
