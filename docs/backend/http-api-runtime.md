# 后端 API 与运行时

## 模块职责

`backend/internal/httpapi/` 负责对外 HTTP 接口、请求参数校验、模型默认值应用、Provider 调度、任务轮询与响应拼装。

## 路由入口

主路由在 `backend/internal/httpapi/router.go`，核心接口如下：

- `GET /healthz`
- `GET /api/meta/images`
- `GET /api/meta/videos`
- `GET /api/meta/audios`
- `POST /api/images/generate`
- `POST /api/audios/generate`
- `POST /api/videos/jobs`
- `GET /api/videos/jobs/{id}`
- `POST /api/animations/jobs`
- `GET /api/animations/jobs/{id}`
- `GET /api/history`
- `GET /api/history/{id}`
- `DELETE /api/history/{id}`
- `GET /api/assets/{id}`

## 请求处理通用模式

1. Handler 解码 JSON
2. 根据 `provider` 和 `model` 补默认值
3. 根据 `models.json` 校验必填字段
4. 通过运行时注册表找到对应 Provider
5. 调用 Provider 并处理结果
6. 需要时把结果交给资产系统归档

相关文件：

- 默认值和校验：`backend/internal/httpapi/model_form_apply.go`, `backend/internal/httpapi/model_form_missing.go`, `backend/internal/httpapi/model_requirements.go`
- Provider 运行时查找：`backend/internal/httpapi/providers_runtime.go`, `backend/internal/httpapi/providers_video_runtime.go`, `backend/internal/httpapi/providers_audio_runtime.go`

## 图片 / 音频 / 视频 Handler

### 图片

- 入口：`backend/internal/httpapi/images_generate.go`
- 特点：通常是同步结果，但仍可能写资产归档

### 音频

- 入口：`backend/internal/httpapi/audios_generate.go`
- 特点：
  - 支持本地结果入库
  - 支持远程 URL 抓取后再入库
  - 最终统一改写为 `/api/assets/{id}`

### 视频

- 创建任务：`backend/internal/httpapi/videos_create.go`
- 查询结果：`backend/internal/httpapi/videos_get.go`
- 特点：
  - 前端拿到 `job_id` 后自行轮询
  - 后端查询下游结果后再决定是否写历史资产

## 日志排查

下游请求日志统一在 `backend/internal/logging/downstream.go`：

- `DownstreamRequest`
- `DownstreamRequestDebug`
- `DownstreamResponse`
- `DownstreamResponseDebug`

当前日志支持记录：

- Provider
- method
- 脱敏后的 URL
- 参数摘要
- 响应状态码
- 耗时
- 下游返回体摘要

相关环境变量：

- `LOG_PROMPT_FULL`
- `LOG_PROMPT_PREVIEW_CHARS`
- `LOG_DOWNSTREAM_BODY_CHARS`
- `LOG_DOWNSTREAM_BODY_FULL`

## 什么时候改这里

- 新增接口或改接口返回结构
- 某模型必填字段没有生效
- 前端报错但 Provider 本身是通的
- 下游日志不足以定位问题
