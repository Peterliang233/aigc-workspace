# Provider 与模型配置

## 模块职责

这个模块负责两件事：

1. 定义“前端能看到哪些平台和模型”
2. 把请求路由到具体平台实现

## 模型配置

统一入口：`backend/models.json`

这个文件决定：

- 默认 provider
- 默认 model
- 模型是否属于 `image` / `video` / `audio`
- 前端表单字段
- 字段默认值、必填项、下拉选项
- 某些模型是否 `requires_image`

如果一个需求是“前端要多展示一个参数”或“某模型要多一个选项”，优先改这里。

## Provider 运行时注册

统一入口：`backend/internal/httpapi/providers_build.go`

这里根据 `.env` 里的配置构建 Provider 实例，并分别注册到：

- `imageProviders`
- `videoProviders`
- `audioProviders`

当前已接入的平台包括：

- `mock`
- `siliconflow`
- `openai_compatible`
- `bltcy`
- `gpt_best`
- `wuyinkeji`
- `jeniya`

## Provider 代码位置

- `backend/internal/providers/siliconflow/`
- `backend/internal/providers/openai_compatible/`
- `backend/internal/providers/gptbest/`
- `backend/internal/providers/wuyinkeji/`
- `backend/internal/providers/jeniya/`
- `backend/internal/providers/mock/`

每个 Provider 至少会实现其中一部分能力：

- 图片生成
- 视频任务创建 / 查询
- 音频生成

## 新增 Provider 的标准步骤

1. 在 `backend/internal/providers/<provider>/` 创建实现
2. 在 `backend/internal/httpapi/providers_build.go` 注册运行时实例
3. 在 `backend/models.json` 增加平台和模型定义
4. 视情况补 `meta` 表单字段
5. 验证前端是否能拿到对应模型

## 新增模型的标准步骤

1. 找到对应 provider 的 `image` / `video` / `audio` 配置块
2. 加入模型 ID、标签、表单字段
3. 如果需要图片输入，设置 `requires_image`
4. 如果有字段映射差异，在具体 provider 里补参数转换逻辑

## 常见排查点

### 前端看不到新模型

先查：

- `backend/models.json`
- `backend/internal/httpapi/meta_images.go`
- `backend/internal/httpapi/meta_videos.go`
- `backend/internal/httpapi/meta_audios.go`

### 参数传了但下游没收到

先查：

- `frontend/src/components/form/ModelFields.tsx`
- `frontend/src/components/video/videoRequest.ts`
- `backend/internal/httpapi/model_form_apply.go`
- 对应 provider 的 payload 构造逻辑

### 某个平台只有图片没有视频

先查：

- `backend/internal/httpapi/providers_build.go`
- 对应 provider 目录是否实现了 `StartVideoJob` / `GetVideoJob`

## 建议修改边界

- 只改表单展示，不改 Provider：优先改 `backend/models.json`
- 只改平台协议，不改页面：优先改 `backend/internal/providers/*`
- 同时新增字段透出和下游映射：`models.json` + provider 一起改
