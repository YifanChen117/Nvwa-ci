# 女娲CI SDK

面向版本控制系统（VCS）的轻量 Go SDK，遵循最小可用性、最少依赖、安全稳定与易扩展原则。当前内置 GitLab Provider，统一提供分支、合并请求（MR）、流水线、作业、提交等能力。

## 目标与原则

- 最小可用性：仅暴露必要 API 与类型，避免无用实体。
- 最少依赖：除标准库，仅依赖 `github.com/xanzy/go-gitlab`。
- 安全稳定：错误显式返回；支持 `context.Context`；无 panic；可单元测试。
- 易扩展：通过 `VCSProvider` 接口实现插件式扩展（GitHub/Bitbucket 等）。
- 统一规范：方法命名统一（Create/List/Get/Accept），类型统一在 `sdk/types`。

## 模块结构

- `sdk/types`：统一领域类型，屏蔽第三方类型差异。
- `sdk/provider`：抽象 `VCSProvider` 接口，定义能力边界。
- `sdk/provider/gitlab`：GitLab Provider 的具体实现（使用 go-gitlab）。
- `sdk/client`：面向上层的统一客户端封装与便捷构造。

代码参考：

- 接口定义：`sdk/provider/provider.go:8`
- 类型定义：`sdk/types/types.go:3`
- 客户端构造：`sdk/client/client.go:13`、`sdk/client/gitlab.go:7`
- GitLab 实现：`sdk/provider/gitlab/gitlab.go:27`

## 安装与集成

- Go 版本：建议 Go 1.21（参考根项目 `go.mod`）。
- 依赖：`github.com/xanzy/go-gitlab v0.115.0`（根项目已包含）。
- 本仓库内使用（当前模块名为 `webci-refactored`）：
  - 导入路径示例：`webci-refactored/sdk/client`
- 发布到 GitHub（推荐独立仓库）：
  - 在新仓库的 `go.mod` 设置 `module github.com/<yourname>/webci-sdk`
  - 其他项目通过 `go get github.com/<yourname>/webci-sdk@v0.1.0` 引用

## 快速开始

环境变量（示例）：

```
GITLAB_BASE_URL="gitlab.oscd"
GITLAB_TOKEN=<your_token>
GITLAB_PROJECT_ID=group/project   # 或项目数字 ID
```

客户端初始化：

```go
import (
  "context"
  "os"
  "webci-refactored/sdk/client"
  "webci-refactored/sdk/types"
)

func example() error {
  c, err := client.NewGitLabClient(os.Getenv("GITLAB_TOKEN"), os.Getenv("GITLAB_BASE_URL"), os.Getenv("GITLAB_PROJECT_ID"))
  if err != nil { return err }
  ctx := context.Background()

  // 创建分支
  _, err = c.CreateBranch(ctx, "feature/sdk-doc", "main")
  if err != nil { return err }

  // 创建 MR
  mr, err := c.CreateMergeRequest(ctx, types.CreateMRInput{
    SourceBranch: "feature/sdk-doc",
    TargetBranch: "main",
    Title:        "SDK Doc MR",
    Description:  "Update docs",
    Squash:       true,
    RemoveSource: true,
  })
  if err != nil { return err }

  // 轮询直到可合并（checking/unchecked → can_be_merged）
  // 参考逻辑实现：`internal/logic/gitlab/gitlab.go:240`
  for {
    cur, err := c.GetMergeRequest(ctx, mr.IID)
    if err != nil { return err }
    if cur.MergeStatus == "can_be_merged" { break }
    time.Sleep(1500 * time.Millisecond)
  }

  // 接受 MR
  _, err = c.AcceptMergeRequest(ctx, mr.IID, types.AcceptMROptions{
    Squash:                    true,
    RemoveSourceBranch:        true,
    MergeWhenPipelineSucceeds: false,
    MergeCommitMessage:        "merge by sdk",
  })
  return err
}
```

## API 参考

### Client（统一入口）

- 构造：
  - `New(provider)`：`sdk/client/client.go:13`
  - `NewGitLabClient(token, baseURL, projectID)`：`sdk/client/gitlab.go:7`
- 分支：
  - `CreateBranch(ctx, name, baseRef)`：`sdk/client/client.go:17`
  - `ListBranches(ctx)`：`sdk/client/client.go:21`
- 合并请求（MR）：
  - `CreateMergeRequest(ctx, CreateMRInput)`：`sdk/client/client.go:25`
  - `GetMergeRequest(ctx, iid)`：`sdk/client/client.go:33`
  - `AcceptMergeRequest(ctx, iid, AcceptMROptions)`：`sdk/client/client.go:29`
- 流水线与作业：
  - `ListPipelines(ctx, PipelineListOptions)`：`sdk/client/client.go:37`
  - `ListJobs(ctx, pipelineID)`：`sdk/client/client.go:41`
- 提交：
  - `GetCommit(ctx, sha)`：`sdk/client/client.go:45`

## 完整 API 接口

### 方法签名（Client）

```go
// 构造
func New(provider provider.VCSProvider) *Client                                  // sdk/client/client.go:13
func NewGitLabClient(token, baseURL, projectID string) (*Client, error)         // sdk/client/gitlab.go:7

// 分支
func (c *Client) CreateBranch(ctx context.Context, name, baseRef string) (*types.Branch, error) // sdk/client/client.go:17
func (c *Client) ListBranches(ctx context.Context) ([]*types.Branch, error)                      // sdk/client/client.go:21

// 合并请求（MR）
func (c *Client) CreateMergeRequest(ctx context.Context, in types.CreateMRInput) (*types.MergeRequest, error) // sdk/client/client.go:25
func (c *Client) GetMergeRequest(ctx context.Context, iid int) (*types.MergeRequest, error)                   // sdk/client/client.go:33
func (c *Client) AcceptMergeRequest(ctx context.Context, iid int, opts types.AcceptMROptions) (*types.MergeRequest, error) // sdk/client/client.go:29

// 流水线与作业
func (c *Client) ListPipelines(ctx context.Context, opts types.PipelineListOptions) ([]*types.Pipeline, error) // sdk/client/client.go:37
func (c *Client) ListJobs(ctx context.Context, pipelineID int) ([]*types.Job, error)                            // sdk/client/client.go:41

// 提交
func (c *Client) GetCommit(ctx context.Context, sha string) (*types.Commit, error) // sdk/client/client.go:45
```

### 方法签名（Provider 接口）

```go
type VCSProvider interface { // sdk/provider/provider.go:8
    CreateBranch(ctx context.Context, name string, baseRef string) (*types.Branch, error)
    ListBranches(ctx context.Context) ([]*types.Branch, error)
    CreateMergeRequest(ctx context.Context, in types.CreateMRInput) (*types.MergeRequest, error)
    AcceptMergeRequest(ctx context.Context, iid int, opts types.AcceptMROptions) (*types.MergeRequest, error)
    GetMergeRequest(ctx context.Context, iid int) (*types.MergeRequest, error)
    ListPipelines(ctx context.Context, opts types.PipelineListOptions) ([]*types.Pipeline, error)
    ListJobs(ctx context.Context, pipelineID int) ([]*types.Job, error)
    GetCommit(ctx context.Context, sha string) (*types.Commit, error)
}
```

### GitLab Provider 实现（主要方法对照）

- 构造：`New(token, baseURL, projectID)`（`sdk/provider/gitlab/gitlab.go:17`）
- 分支：`CreateBranch`（`sdk/provider/gitlab/gitlab.go:27`）、`ListBranches`（`sdk/provider/gitlab/gitlab.go:36`）
- MR：`CreateMergeRequest`（`sdk/provider/gitlab/gitlab.go:48`）、`GetMergeRequest`（`sdk/provider/gitlab/gitlab.go:82`）、`AcceptMergeRequest`（`sdk/provider/gitlab/gitlab.go:66`）
- 流水线：`ListPipelines`（`sdk/provider/gitlab/gitlab.go:90`）
- 作业：`ListJobs`（`sdk/provider/gitlab/gitlab.go:106`）
- 提交：`GetCommit`（`sdk/provider/gitlab/gitlab.go:118`）
- MR 映射：`toMR`（`sdk/provider/gitlab/gitlab.go:126`）

### 类型定义（Types）

```go
// 分支 sdk/types/types.go:3
type Branch struct {
    Name      string
    CommitSHA string
    Protected bool
}

// 合并请求 sdk/types/types.go:9
type MergeRequest struct {
    IID                          int
    State                        string
    Title                        string
    SourceBranch                 string
    TargetBranch                 string
    MergeStatus                  string
    HasConflicts                 bool
    WorkInProgress               bool
    BlockingDiscussionsResolved  bool
}

// 流水线 sdk/types/types.go:21
type Pipeline struct {
    ID     int
    Status string
    Ref    string
    SHA    string
    WebURL string
}

// 作业 sdk/types/types.go:29
type Job struct {
    ID     int
    Name   string
    Status string
    Stage  string
}

// 提交 sdk/types/types.go:36
type Commit struct {
    ID          string
    ShortID     string
    Title       string
    Message     string
    AuthorName  string
    AuthorEmail string
    CreatedAt   string
}

// 创建 MR 入参 sdk/types/types.go:46
type CreateMRInput struct {
    SourceBranch string
    TargetBranch string
    Title        string
    Description  string
    Squash       bool
    RemoveSource bool
    MWPS         bool
}

// 接受 MR 选项 sdk/types/types.go:56
type AcceptMROptions struct {
    Squash                    bool
    RemoveSourceBranch        bool
    MergeWhenPipelineSucceeds bool
    MergeCommitMessage        string
}

// 流水线列表参数 sdk/types/types.go:63
type PipelineListOptions struct {
    Page    int
    PerPage int
}
```

### 字段语义与约束

- `Branch.Name`：分支名称；`CommitSHA`：分支最新提交；`Protected`：是否受保护。
- `MergeRequest.MergeStatus`：可合并状态（示例：`can_be_merged`、`checking`、`unchecked` 等）。
- `AcceptMROptions.RemoveSourceBranch`：接受 MR 后是否删除源分支。
- `AcceptMROptions.MergeWhenPipelineSucceeds`：流水线成功后自动合并（需项目启用相关策略）。
- `PipelineListOptions.Page/PerPage`：分页参数；不设则使用默认。


### Provider 接口（可扩展点）

- 位置：`sdk/provider/provider.go:8`
- 方法：CreateBranch / ListBranches / CreateMergeRequest / AcceptMergeRequest / GetMergeRequest / ListPipelines / ListJobs / GetCommit

### 类型（统一定义）

- 位置：`sdk/types/types.go:3`
- 主要类型：`Branch`、`MergeRequest`、`Pipeline`、`Job`、`Commit`
- 入参类型：`CreateMRInput`、`AcceptMROptions`、`PipelineListOptions`

## GitLab Provider 说明

- 构造：`sdk/provider/gitlab/gitlab.go:17`
- 主要方法：
  - 分支：`CreateBranch`、`ListBranches`（`sdk/provider/gitlab/gitlab.go:27/36`）
  - MR：`CreateMergeRequest`、`GetMergeRequest`、`AcceptMergeRequest`（`sdk/provider/gitlab/gitlab.go:48/82/66`）
  - 流水线/作业/提交：`ListPipelines`、`ListJobs`、`GetCommit`（`sdk/provider/gitlab/gitlab.go:90/106/118`）
- MR 映射：`toMR`（`sdk/provider/gitlab/gitlab.go:126`）

## 错误处理与上下文

- 所有方法返回 `error`；不可合并、冲突、WIP、未解决讨论等由调用者决定如何提示。
- 支持 `context.Context`，建议为长时间操作（MR 轮询）设置超时：

```go
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()
```




