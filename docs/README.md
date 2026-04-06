# Docs 索引

这个目录用于沉淀系统设计与代码入口索引，目标不是写成长篇说明书，而是让后续维护者或模型能快速定位“改哪个模块、看哪些文件、影响哪些链路”。

## 建议阅读顺序

1. `docs/change-map.md`
2. `docs/architecture/system-overview.md`
3. 按功能进入对应模块文档

## 文档目录

- `docs/change-map.md`
  - 需求类型到代码模块的快速索引
- `docs/architecture/system-overview.md`
  - 系统总览、运行链路、核心依赖
- `docs/backend/http-api-runtime.md`
  - 后端 API、Handler、运行时编排
- `docs/backend/providers-models.md`
  - 模型配置、Provider 构建、平台接入方式
- `docs/backend/assets-history.md`
  - MinIO、历史记录、资产归档链路
- `docs/frontend/workbench-ui-state.md`
  - 前端布局、Studio 组件、状态轮询
- `docs/features/animation-workshop.md`
  - 动画工坊、分镜规划、分段生成与拼接

## 快速定位

| 场景 | 先看文档 | 再看代码 |
| --- | --- | --- |
| 新接入一个图片/视频/音频平台 | `docs/backend/providers-models.md` | `backend/internal/httpapi/providers_build.go`, `backend/models.json`, `backend/internal/providers/*` |
| 某个生成接口报错 | `docs/backend/http-api-runtime.md` | `backend/internal/httpapi/*`, 对应 `backend/internal/providers/*` |
| 历史记录查不到数据 | `docs/backend/assets-history.md` | `backend/internal/assets/*`, `backend/internal/httpapi/history.go` |
| 前端表单参数没有透出 | `docs/frontend/workbench-ui-state.md` | `frontend/src/hooks/*Meta.ts`, `frontend/src/components/*Studio.tsx`, `frontend/src/components/form/ModelFields.tsx` |
| 动画工坊链路异常 | `docs/features/animation-workshop.md` | `backend/internal/httpapi/animations_*.go`, `backend/internal/animation/*`, `frontend/src/components/animation/*` |
| 视频/音频结果前端不刷新 | `docs/frontend/workbench-ui-state.md` | `frontend/src/state/generation.tsx`, `frontend/src/state/animation.tsx` |
| 生成结果没有入库或没进 MinIO | `docs/backend/assets-history.md` | `backend/internal/httpapi/images_generate.go`, `backend/internal/httpapi/videos_get.go`, `backend/internal/httpapi/audios_generate.go`, `backend/internal/assets/*` |

## 维护原则

- 新增功能时，至少补一条到 `docs/change-map.md`
- 新增系统级模块时，新增独立文档，不要把所有内容堆到一个文件
- 文档描述以“模块职责 + 入口文件 + 改动影响面”为主，不写与代码脱节的空泛说明
