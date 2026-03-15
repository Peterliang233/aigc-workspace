# AIGC Workspace (React + Go)

目标：把大模型厂商的“生成图片 / 生成视频”能力，封装成一个前后端分离的 Web 应用。

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

### 1) 启动后端

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

后端默认监听 `http://localhost:8080`，健康检查 `GET /healthz`。

如果你在受限的沙箱环境里遇到 Go 构建缓存权限问题，可以临时指定：

```bash
env GOCACHE=/tmp/gocache go run ./cmd/server
```

### 2) 启动前端

```bash
cd frontend
npm i
npm run dev
```

前端默认 `http://localhost:5173`，已在 `vite.config.ts` 里把 `/api` 代理到后端。

## Docker Compose

### 生产模式（构建镜像 + Nginx 托管前端）

```bash
docker compose up -d --build
```

- 前端：`http://localhost:8081`（Nginx，已反代 `/api` `/static` `/healthz` 到后端）
- 后端：`http://localhost:8080`

### 开发模式（容器跑 go run + vite，挂载源码）

```bash
docker compose -f docker-compose.dev.yml up -d --build
docker compose -f docker-compose.dev.yml logs --tail=200 -f
```

前端 `http://localhost:5173`，后端 `http://localhost:8080`。

## API 概览

- `POST /api/images/generate`
  - 入参：`{ "prompt": "...", "size": "1024x1024", "n": 1 }`
  - 出参：`{ "image_urls": ["..."], "provider": "..." }`
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

视频接口因为不同厂商差异很大，默认实现为“可用的异步任务链路 + 可配置的 start/status endpoint”。你可以先跑通 UI，再按目标厂商的 API 协议补齐 Provider 的字段映射。
