# AIGC Workspace

目标：把大模型厂商的“生成图片 / 生成视频”能力，封装成一个 Web 应用，自己直接对接模型厂商，做定制化能力，降低AIGC能力使用成本

## 目录结构

- `backend/`: Go HTTP API（不依赖第三方 Web 框架，便于二次封装）
- `frontend/`: React Web（Vite）

## 本地启动（开发模式）

### 0) Makefile（推荐）

在项目根目录：

```bash
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

## 网页配置（推荐）

前端侧边栏新增了「配置」模块，用于管理不同平台的“模型列表”（新增/删除）。

- Base URL / API Key / 默认模型：通过部署环境配置（env）
- 模型列表：仅存到 MySQL（需要配置 `MYSQL_DSN`；已支持新增/删除）

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

- 没配 Key 时，默认走 `mock` provider：图片会生成一个可访问的 SVG 占位图（用于前后端联调）。
- 配置了 `AIGC_PROVIDER=openai_compatible` 且提供 `AIGC_API_KEY` 时，会走 OpenAI 兼容图片接口：
  - `POST {AIGC_BASE_URL}/v1/images/generations`
  - 当前实现优先使用 `b64_json`，后端会落盘到 `backend/var/generated/` 并通过 `/static/` 暴露
- 配置了 `AIGC_PROVIDER=siliconflow` 时，会调用 SiliconFlow 图片接口并将返回的临时 URL 下载落盘：
  - `POST https://api.siliconflow.cn/v1/images/generations`
  - SiliconFlow 返回的图片 URL 通常是临时有效（约 1 小时），后端会立即下载到 `backend/var/generated/` 并返回本站 `/static/generated/...`，保证网页可稳定展示
- 支持多平台并存：前端会调用 `GET /api/meta/images` 获取可选的平台与模型，并在 `POST /api/images/generate` 时携带 `provider/model`，后端按请求动态路由到对应平台。

视频接口因为不同厂商差异很大，默认实现为“可用的异步任务链路 + 可配置的 start/status endpoint”。你可以先跑通 UI，再按目标厂商的 API 协议补齐 Provider 的字段映射。

### 速创API（无印科技）模型名说明

速创API 的图片生成接口是异步的，且模型名在 URL 中作为动态段出现：

- `POST /api/async/{model}?key=你的密钥`
- 例如 `model=image_nanoBanana_pro`

因此在网页「配置」页里填写的模型列表，建议直接使用 `image_nanoBanana_pro` 这类模型名（而不是完整 URL）。
