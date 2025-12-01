# Nvwa-CI 技术总体设计与实现说明

## 目标与范围

- 提供以 GitLab 为核心的分支与合并请求（MR）管理、CI 可视化页面、任务类型判别与自动合并能力。
- 支持最小依赖、统一规范、可扩展（通过 SDK 的 Provider 接口）、安全稳定的工程实践。
- 交付形式：后端服务（基于 Hertz），以及可独立引用的 VCS SDK 模块（Go 包）。

## 架构总览

- 分层结构：Router → Handler → Logic → Service → 外部 API（GitLab）
- UI 页面由 Handler 输出 HTML 模板，后端通过逻辑层聚合 GitLab 数据供前端展示与操作。

### 路由与入口

- 服务入口：`cmd/server/main.go`
- 路由注册：`internal/router/router.go`
  - 首页路由：`internal/router/router.go:176` 通过 `indexPageHandler` 渲染 GitLab 页面或首页。
  - GitLab API 路由：配置获取与修改、分支与 MR 操作、自动合并等。
  - 自动合并：`internal/router/router.go:170` 注册 `POST /api/gitlab/merge_requests/auto`。

### 处理与页面

- GitLab 处理器：`internal/handler/gitlab/gitlab.go`
  - 页面渲染：`Page`/`PageContent` 输出 CI 页面 HTML。
  - 自动合并入口：`AutoMerge`（参数校验、调用逻辑层创建与接受 MR）。
- 页面模板：`internal/handler/gitlab/template.go`（按钮、自动刷新、任务类型列等）。

### 逻辑层（GitLab）

- 文件：`internal/logic/gitlab/gitlab.go`
- 关键能力：
  - MR 接受前的可合并性轮询与约束检查：`internal/logic/gitlab/gitlab.go:226`；状态轮询在 `internal/logic/gitlab/gitlab.go:240`（处理 `checking/unchecked/cannot_be_merged_recheck`）。
  - 接受 MR 并记录：`internal/logic/gitlab/gitlab.go:261` 调用 Service；`internal/logic/gitlab/gitlab.go:265` 记录合并。
  - CI 页面数据聚合：流水线 → 作业 → 提交信息与任务类型，见 `internal/logic/gitlab/gitlab.go:270` 起。
  - 任务类型判别：对 `GitLabJobInfo` 应用分类，`internal/logic/gitlab/gitlab.go:399` 构造初始记录后，`internal/logic/gitlab/gitlab.go:418` 调用分类方法。

### 服务层（GitLab API 封装）

- 文件：`internal/service/gitlab/gitlab.go`
- 关键方法：
  - 获取分支：`internal/service/gitlab/gitlab.go:165`
  - 获取提交：`internal/service/gitlab/gitlab.go:180` 与 `internal/service/gitlab/gitlab.go:195`
  - 创建分支：`internal/service/gitlab/gitlab.go:217`
  - 创建 MR：`internal/service/gitlab/gitlab.go:230`
  - 接受 MR：`internal/service/gitlab/gitlab.go:285`
  - 流水线与作业：`internal/service/gitlab/gitlab.go:95`、`internal/service/gitlab/gitlab.go:150`

## 功能模块

- 分支管理
  - 创建分支（支持从指定基准 `Ref` 创建）：Service 层 `CreateBranch`；前端按钮与自动刷新在模板中实现。
  - 分支前缀规范：如 `feature/`、`test/`、`release/`，由逻辑层与 UI 共同校验与引导（可在后续统一后端校验）。
- 合并请求（MR）
  - 创建 MR：输入源/目标分支、标题、描述（可选）、Squash/删除源分支（可选）。
  - 自动合并：后端轮询合并状态后调用 `AcceptMergeRequest`；合并失败返回明确原因，提示手动处理。
- CI 页面
  - 展示流水线与作业列表，标签化任务类型（创建分支/分支合并/修改文件），并显示提交信息与触发用户等。
- 任务类型判别
  - 基于 GitLab API 数据（SHA、MR 状态）进行判别，避免仅前端策略导致重启后丢失类型。

## SDK 设计

- 模块：`sdk/types`、`sdk/provider`、`sdk/provider/gitlab`、`sdk/client`
- 原则：最小可用性与最少依赖；统一类型在 `sdk/types`，通过 `VCSProvider` 插件式扩展。
- 快速构造：`sdk/client/gitlab.go:7` `NewGitLabClient(token, baseURL, projectID)`
- 完整 API：详见 `sdk/README.md` 的“完整 API 接口”章节（Client/Provider 方法签名、类型定义、字段语义）。

## 配置与环境变量

- 位置：`internal/config/config.go`
- 环境变量：
  - `HTTP_ADDR`（默认 `:8080`）
  - `MYSQL_DSN`（留空使用内存 SQLite 演示）
  - `REPO_PATH`（用于分支 refresh 功能，可选）
  - `GITLAB_BASE_URL`（例如 `https://gitlab.example.com/api/v4`）
  - `GITLAB_TOKEN`（访问令牌）
  - `GITLAB_PROJECT_ID`（项目路径或数字 ID）

## 安全与稳定性

- 错误处理：所有接口返回 `error`，不使用 `panic`；统一错误消息返回到前端。
- 上下文传递：SDK 与 Provider 方法支持 `context.Context`，便于超时与取消。
- 证书与网络：GitLab Provider 默认使用不安全 TLS 以支持本地自签证书测试，生产建议改为严格 TLS，并在后续支持注入自定义 `http.Client`。
- 机密管理：令牌与项目 ID 使用环境变量注入；避免硬编码或提交到仓库。


## 部署与运行

- 运行服务：`go run ./cmd/server`（或 `make run`，参考 `Makefile`）
- 访问页面：`http://localhost:8080/`（首页）与 `/gitlab`（CI 页面）
- 配置变更：通过 `POST /api/gitlab/config` 修改 GitLab BaseURL/Token/ProjectID/RepoPath。


---

## 代码索引（快速跳转）

- 路由：`internal/router/router.go:170`、`internal/router/router.go:176`
- 处理器：`internal/handler/gitlab/gitlab.go:59`
- 逻辑层：
  - MR 轮询与接受：`internal/logic/gitlab/gitlab.go:226`、`internal/logic/gitlab/gitlab.go:240`、`internal/logic/gitlab/gitlab.go:261`、`internal/logic/gitlab/gitlab.go:265`
  - CI 聚合与类型判别：`internal/logic/gitlab/gitlab.go:270`、`internal/logic/gitlab/gitlab.go:399`、`internal/logic/gitlab/gitlab.go:418`
- 服务层：
  - 分支/提交：`internal/service/gitlab/gitlab.go:165`、`internal/service/gitlab/gitlab.go:180`、`internal/service/gitlab/gitlab.go:195`
  - 创建分支/MR/接受 MR：`internal/service/gitlab/gitlab.go:217`、`internal/service/gitlab/gitlab.go:230`、`internal/service/gitlab/gitlab.go:285`
- SDK：
  - 类型：`sdk/types/types.go`
  - Provider 接口：`sdk/provider/provider.go:8`
  - GitLab Provider：`sdk/provider/gitlab/gitlab.go:17`、`sdk/provider/gitlab/gitlab.go:27`、`sdk/provider/gitlab/gitlab.go:48`、`sdk/provider/gitlab/gitlab.go:66`、`sdk/provider/gitlab/gitlab.go:82`
  - Client 封装：`sdk/client/client.go:13`、`sdk/client/gitlab.go:7`

