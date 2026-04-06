# 资产归档与历史记录

## 模块职责

系统把生成出来的图片、视频、音频、动画结果统一沉淀到资产系统，目的是：

- 避免直接依赖平台临时 URL
- 支撑历史记录查询
- 支撑统一下载、预览、播放

## 核心模块

- 资产服务：`backend/internal/assets/assets.go`
- MySQL 存储：`backend/internal/assets/mysql_store.go`
- 本地文件入库：`backend/internal/assets/service_local.go`
- 远程 URL 抓取入库：`backend/internal/assets/service_remote.go`
- MinIO 封装：`backend/internal/blobstore/*`

## 资产写入来源

### 图片

- `backend/internal/httpapi/images_generate.go`

### 视频

- `backend/internal/httpapi/videos_get.go`
- 视频成功后通常在查询结果阶段归档

### 音频

- `backend/internal/httpapi/audios_generate.go`
- 既支持本地文件归档，也支持远程 URL 拉取后归档

### 动画

- `backend/internal/httpapi/animations_media.go`
- 最终拼接结果完成后归档

## 历史接口

入口文件：`backend/internal/httpapi/history.go`

支持：

- 列表查询
- 按能力筛选
- 关键字搜索
- 分页
- 详情查询

相关接口：

- `GET /api/history`
- `GET /api/history/{id}`
- `DELETE /api/history/{id}`
- `GET /api/assets/{id}`

## 数据结构关键字段

历史记录会保留这些核心信息：

- `capability`
- `provider`
- `model`
- `status`
- `error`
- `prompt_preview`
- `content_type`
- `bytes`
- `url`
- `created_at`

因此如果历史页缺失“模型”或“平台”，优先从资产写入链路检查是否落库。

## 常见问题定位

### 前端有生成结果，但历史里没有

先查：

- 是否走到了资产写入分支
- 是否因为 MinIO 未启用导致资产系统不可用
- 是否只返回了下游 URL，未执行 `StoreRemote`

### 历史有记录，但资源打不开

先查：

- `backend/internal/httpapi/assets_stream.go`
- MinIO 对象是否存在
- 资产记录里的 `content_type`

### 模型字段没透出到历史页

先查：

- 资产写入时是否把 `Model` 带入
- `backend/internal/httpapi/history.go` 是否序列化了该字段
- `frontend/src/components/history/HistoryRow.tsx` 是否展示了该字段
