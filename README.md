# AIGC Workspace

目标：把大模型厂商的“生成图片 / 生成视频”能力，封装成一个 Web 应用，自己直接对接模型厂商，做定制化能力，降低AIGC能力使用成本

## 目录结构

- `backend/`: Go HTTP API（不依赖第三方 Web 框架，便于二次封装）
- `frontend/`: React Web（Vite）

## 本地启动（开发模式）

### 0) Makefile（推荐）

在项目根目录：

```bash
make mysql-up
make minio-up
make backend-test
make -j2 dev
```

注意：根目录 `.env` 是给 Docker Compose 和后端的 `.env` loader 使用的，不要用 `source .env` 的方式加载到 shell 里（例如 MySQL DSN 里的 `tcp(127.0.0.1:3307)` 不是合法的 POSIX shell 语法）。

### 1) 启动后端

```bash
cp .env.example .env
cd backend
env GOCACHE=/tmp/gocache go run .
```

后端默认监听 `http://localhost:8080`，健康检查 `GET /healthz`。

如果你在受限的沙箱环境里遇到 Go 构建缓存权限问题，可以临时指定：

```bash
env GOCACHE=/tmp/gocache go run .
```

### 2) 启动前端

```bash
cd frontend
npm i
npm run dev
```

前端默认 `http://localhost:5173`，已在 `vite.config.ts` 里把 `/api` 代理到后端。

## MinIO 说明（历史资源存储）

MinIO 用于保存生成结果（图片/视频）的归档副本，避免直接依赖下游厂商 URL（可能过期或不可访问），并支撑历史记录回显。

### 快速启动（开发）

```bash
cp .env.example .env
make minio-up
```

等价命令：

```bash
docker compose -f docker-compose.dev.yml up -d minio minio-init
```

- API Endpoint（宿主机）：`127.0.0.1:9000`（默认）
- Console：`http://localhost:9001`
- 默认账号密码：`minioadmin / minioadmin`（来自 `.env` 的 `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD`）
- 默认 Bucket：`aigc-assets`（`MINIO_BUCKET`，`minio-init` 和后端都会确保 bucket 存在）

### 配置说明

- `MINIO_ENDPOINT=minio:9000`：容器网络内访问地址（Docker Compose 内部服务互联）。
- `MINIO_ENDPOINT_LOCAL=127.0.0.1:9000`：后端本机 `go run` 时的回退地址。
- 后端会自动判断当前环境：如果无法解析 `minio` 主机名，会自动使用 `MINIO_ENDPOINT_LOCAL`。
- `MINIO_USE_SSL=false`：本地开发默认关闭 TLS；生产建议开启 TLS 并使用受信证书。

### 不使用 MinIO（仅临时调试）

若你只想先跑通接口，可将 `.env` 里的 `MINIO_ENDPOINT` 和 `MINIO_ENDPOINT_LOCAL` 都置空。后端会以 `minio_disabled` 模式启动，但历史资源归档能力不可用。

## 模型与表单配置（JSON）

不同平台的不同模型，可能需要不同的必填表单（例如 I2V 模型需要参考图）。本项目将「平台/模型列表 + 表单要求」统一存储在 `backend/models.json` 中，由后端在启动时加载，并通过 `GET /api/meta/images`、`GET /api/meta/videos` 下发给前端驱动 UI。

- 平台 Base URL / API Key：只通过 `.env` 配置（不进库）
- 模型列表 + 表单映射：只通过 `backend/models.json` 配置（不进库、不可在网页编辑）

## Docker Compose

### 生产模式（构建镜像 + Nginx 托管前端）

```bash
cp .env.example .env
docker compose up -d --build
```

- 前端：`http://localhost:8081`（Nginx，已反代 `/api` `/static` `/healthz` 到后端）
- 后端：`http://localhost:8080`

### 开发模式（容器跑 go run + vite，挂载源码）

```bash
cp .env.example .env
docker compose -f docker-compose.dev.yml up -d --build
docker compose -f docker-compose.dev.yml logs --tail=200 -f
```

前端 `http://localhost:5173`，后端 `http://localhost:8080`。

MySQL 默认会映射到宿主机 `3307`（避免和本机已有的 MySQL `3306` 冲突）。如需修改，编辑根目录 `.env` 里的 `MYSQL_PORT`，并同步调整 `MYSQL_DSN_LOCAL`。
MinIO 默认映射 `9000`（API）和 `9001`（Console）。如需修改，编辑根目录 `.env` 里的 `MINIO_PORT` / `MINIO_CONSOLE_PORT`。

## API 概览

- `POST /api/images/generate`
  - 入参：`{ "provider": "siliconflow", "model": "Kwai-Kolors/Kolors", "prompt": "...", "size": "1024x1024", "n": 1 }`
  - 出参：`{ "image_urls": ["..."], "provider": "...", "model": "..." }`
- `POST /api/videos/jobs`
  - 入参：`{ "prompt": "...", "duration_seconds": 5, "aspect_ratio": "16:9" }`
  - 出参：`{ "job_id": "...", "status": "queued", "provider": "..." }`
- `GET /api/videos/jobs/{id}`
  - 出参：`{ "job_id": "...", "status": "queued|running|succeeded|failed", "video_url": "...", "error": "..." }`

## 厂商接入（Provider）

后端用 Provider 抽象封装了“图片生成”和“视频生成(异步任务)”两类能力：

- 没配 Key 时，可以先用 `mock` provider：图片会生成一个可访问的 SVG 占位图（用于前后端联调）。
- 图片：前端通过 `GET /api/meta/images` 获取可选的平台与模型，并在 `POST /api/images/generate` 时携带 `provider/model`，后端按请求动态路由到对应平台。
- 视频：前端通过 `GET /api/meta/videos` 获取可选的平台与模型。若模型配置为 `requires_image=true`（在 `backend/models.json`），前后端都会强制要求提供参考图片。

视频接口因为不同厂商差异很大，默认实现为“可用的异步任务链路 + 可配置的 start/status endpoint”。你可以先跑通 UI，再按目标厂商的 API 协议补齐 Provider 的字段映射。

### BLTCY（柏拉图）接入说明

已内置 `bltcy` 图片 provider（走 `POST /v1/chat/completions`，默认模型 `gpt-4o-image`）。

在根目录 `.env` 配置：

```bash
BLTCY_BASE_URL=https://api.bltcy.ai
BLTCY_API_KEY=你的密钥
```

重启后端后，前端「图片生成」里可直接选择 `BLTCY(柏拉图)` 并切换模型。

### 速创API（无印科技）模型名说明

速创API 的图片生成接口是异步的，且模型名在 URL 中作为动态段出现：

- `POST /api/async/{model}?key=你的密钥`
- 例如 `model=image_nanoBanana_pro`

因此在 `backend/models.json` 里配置模型列表时，建议直接使用 `image_nanoBanana_pro` 这类模型名（而不是完整 URL）。
