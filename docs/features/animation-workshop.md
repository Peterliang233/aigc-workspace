# 动画工坊

## 目标

动画工坊不是单次视频生成，而是“总时长动画编排器”：

- 输入一个总 Prompt
- 指定总时长
- 自动拆成多个连续镜头段落
- 用图生视频模型逐段生成
- 通过尾帧衔接保证镜头连续性
- 最后拼接成一个完整动画

## 后端链路

### 任务创建

- 入口：`backend/internal/httpapi/animations_create.go`
- 负责校验：
  - provider
  - model
  - 总时长
  - 是否是图生视频模型

### 分段规划

- 规划规则：`backend/internal/animation/plan.go`
- Prompt 规划：`backend/internal/httpapi/animations_planner.go`
- 规划器客户端：`backend/internal/animation/planner_client.go`

当前策略：

1. 先根据目标时长和模型支持时长拆分 segment durations
2. 若规划模型可用，则把总 Prompt 改写为多段连续 Prompt
3. 若规划失败，则退回 fallback segments

### 分段执行

- 执行主链路：`backend/internal/httpapi/animations_run.go`

核心过程：

1. 生成第一段首帧参考图
2. 每段调用视频 Provider 生成视频
3. 轮询该段下游 `sourceJobID`
4. 下载该段视频到临时目录
5. 抽取尾帧作为下一段 `leadImage`

### 媒体处理

- 媒体客户端：`backend/internal/animation/media_client.go`
- HTTP worker 实现：`backend/internal/animation/media_client_http.go`
- 动画媒体辅助 Handler：`backend/internal/httpapi/animations_media.go`

依赖能力：

- `ExtractLastFrame`
- `ConcatAndTrim`

## 前端链路

- 页面入口：`frontend/src/components/AnimationStudio.tsx`
- 表单：`frontend/src/components/animation/AnimationJobForm.tsx`
- 结果：`frontend/src/components/animation/AnimationResults.tsx`
- 状态：`frontend/src/state/animation.tsx`

前端展示的关键状态包括：

- 总任务状态
- planner 状态
- segment 列表
- 每段 prompt
- 每段 source job id
- 中间视频结果
- 最终拼接结果

## 什么时候改这里

### 想优化镜头连续性

先查：

- `backend/internal/animation/planner_prompt.go`
- `backend/internal/httpapi/animations_planner.go`
- `backend/internal/httpapi/animations_run.go`

### 想调整拆段规则

先查：

- `backend/internal/animation/plan.go`

### 想把中间结果更多透到前端

先查：

- `backend/internal/httpapi/animations_get.go`
- `frontend/src/components/animation/AnimationResults.tsx`

### 想替换媒体处理实现

先查：

- `backend/internal/animation/media_client.go`
- `backend/internal/animation/media_client_http.go`
