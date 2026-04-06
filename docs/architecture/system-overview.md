# 系统总览

## 目标

系统把多家 AIGC 平台的图片、视频、音频、动画能力封装成统一 Web 工作台，核心诉求是：

- 前端统一操作入口
- 后端统一模型配置与平台接入
- 生成结果统一归档到 MySQL + MinIO
- 支持历史记录、任务轮询、动画分段生成

## 运行结构

### 后端

- 入口：`backend/main.go`
- 配置加载：`backend/internal/config/config.go`
- 路由组装：`backend/internal/httpapi/router.go`
- 模型配置：`backend/models.json`
- Provider 实现：`backend/internal/providers/*`
- 资产归档：`backend/internal/assets/*`
- MinIO 封装：`backend/internal/blobstore/*`

### 前端

- 应用入口：`frontend/src/App.tsx`
- 主布局：`frontend/src/layout/*`
- Studio 页面：`frontend/src/components/*Studio.tsx`
- 动态模型元数据：`frontend/src/hooks/use*Meta.ts`
- 生成状态管理：`frontend/src/state/generation.tsx`
- 动画状态管理：`frontend/src/state/animation.tsx`

## 核心链路

### 图片 / 音频同步链路

1. 前端提交表单
2. 后端 Handler 解析请求并选择 Provider
3. Provider 调用下游模型接口
4. 若结果需要归档，则写 MySQL / MinIO
5. 前端拿到最终 URL 并展示

### 视频异步链路

1. 前端请求 `POST /api/videos/jobs`
2. 后端创建下游任务并返回 `job_id`
3. 前端通过全局状态轮询 `GET /api/videos/jobs/{id}`
4. Provider 查询下游任务结果
5. 成功后写入资产系统，前端更新播放地址

### 动画工坊链路

1. 用户输入总 Prompt、总时长、模型
2. 后端按模型支持的时长拆分分段计划
3. 规划器把总 Prompt 改写为多个连续镜头 Prompt
4. 每段调用图生视频模型生成
5. 抽尾帧作为下一段首帧参考
6. 媒体工作器拼接并裁剪为最终视频
7. 最终结果归档并回传前端

## 核心设计约束

- Provider 的平台账号、Base URL、API Key 只从 `.env` 加载
- 模型列表与前端表单只从 `backend/models.json` 透出
- 历史记录依赖 MySQL；资源归档依赖 MinIO
- 动画工坊依赖媒体工作器完成抽帧与拼接，不直接依赖宿主机本地工具

## 常见修改入口

- 改接口协议：优先看 `backend/internal/providers/*`
- 改表单透出：优先看 `backend/models.json`
- 改路由或响应结构：优先看 `backend/internal/httpapi/*`
- 改页面布局或交互：优先看 `frontend/src/App.tsx`、`frontend/src/layout/*`、`frontend/src/components/*`
