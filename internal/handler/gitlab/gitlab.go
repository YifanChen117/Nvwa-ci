package gitlab

import (
	"embed"
	"log"
	"strconv"
	"strings"
	"webci-refactored/internal/config"
	"webci-refactored/internal/logic/gitlab"

	"github.com/cloudwego/hertz/pkg/app"
)

var gitlabTemplate embed.FS

// Handler GitLab处理层
type Handler struct {
	logic *gitlab.Logic
}

// NewHandler 创建GitLab处理层实例
func NewHandler(cfg config.Config) (*Handler, error) {
	log.Printf("Creating GitLab handler with config: baseURL=%s, project=%s", cfg.GitLabBaseURL, cfg.GitLabProject)

	// 创建GitLab业务逻辑
	logic, err := gitlab.NewLogic(cfg)
	if err != nil {
		log.Printf("Failed to create gitlab logic: %v", err)
		return nil, err
	}

	return &Handler{
		logic: logic,
	}, nil
}

// Ok 返回成功响应
func Ok(c *app.RequestContext, data interface{}) {
	c.JSON(200, map[string]interface{}{"code": 0, "message": "ok", "data": data})
}

// Err 返回错误响应
func Err(c *app.RequestContext, status int, msg string) {
	log.Printf("GitLab API error: status=%d, message=%s", status, msg)
	c.JSON(status, map[string]interface{}{"code": status, "message": msg})
}

// PageContent 返回GitLab流水线任务页面的HTML内容
func (h *Handler) PageContent() string {
	return GetJobListTemplate()
}

// Page 返回GitLab流水线任务页面
func (h *Handler) Page(c *app.RequestContext) {
	html := h.PageContent()
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.SetStatusCode(200)
	c.Write([]byte(html))
}

// ListPipelines 列出流水线
func (h *Handler) ListPipelines(c *app.RequestContext) {
	log.Printf("Handling ListPipelines request")
	pipelines, err := h.logic.ListPipelines()
	if err != nil {
		log.Printf("Failed to list pipelines: %v", err)
		Err(c, 500, err.Error())
		return
	}

	log.Printf("Successfully listed %d pipelines", len(pipelines))
	Ok(c, pipelines)
}

// GetPipeline 获取流水线详情
func (h *Handler) GetPipeline(c *app.RequestContext) {
	log.Printf("Handling GetPipeline request")
	// 从路径参数读取 id，并转换为整数
	idStr := string(c.Param("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid pipeline id: %s", idStr)
		Err(c, 400, "invalid pipeline id")
		return
	}

	log.Printf("Getting details for pipeline %d", id)
	details, err := h.logic.GetPipelineDetails(id)
	if err != nil {
		log.Printf("Failed to get pipeline %d: %v", id, err)
		Err(c, 500, err.Error())
		return
	}

	log.Printf("Successfully got details for pipeline %d", id)
	Ok(c, details)
}

// ListBranches 列出分支
func (h *Handler) ListBranches(c *app.RequestContext) {
	log.Printf("Handling ListBranches request")
	branches, err := h.logic.ListBranches()
	if err != nil {
		log.Printf("Failed to list branches: %v", err)
		Err(c, 500, err.Error())
		return
	}

	log.Printf("Successfully listed %d branches", len(branches))
	Ok(c, branches)
}

// ListJobs 列出流水线任务（用于CI模拟器页面展示）
func (h *Handler) ListJobs(c *app.RequestContext) {
	log.Printf("Handling ListJobs request for CI simulator page")
	page := 1
	perPage := 20
	if v := c.Query("page"); len(v) > 0 {
		if n, err := strconv.Atoi(string(v)); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("per_page"); len(v) > 0 {
		if n, err := strconv.Atoi(string(v)); err == nil && n > 0 {
			perPage = n
		}
	}
	pageData, err := h.logic.ListJobsPage(page, perPage)
	if err != nil {
		log.Printf("Failed to list jobs for CI page: %v", err)
		Err(c, 500, err.Error())
		return
	}

	log.Printf("Successfully listed %d jobs for CI simulator page", len(pageData.Items))
	Ok(c, pageData)
}

func (h *Handler) UpdateConfig(cfg config.Config) error {
	return h.logic.UpdateConfig(cfg)
}

func (h *Handler) CreateBranch(c *app.RequestContext) {
	var in struct {
		Name string `json:"name"`
		Ref  string `json:"ref"`
	}
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	if in.Name == "" {
		Err(c, 400, "name required")
		return
	}
	if !(strings.HasPrefix(in.Name, "feature/") || strings.HasPrefix(in.Name, "test/") || strings.HasPrefix(in.Name, "release/")) {
		Err(c, 400, "invalid branch prefix")
		return
	}
	if in.Name == "feature/" || in.Name == "test/" || in.Name == "release/" {
		Err(c, 400, "invalid branch name")
		return
	}
	in.Ref = "main"
	b, err := h.logic.CreateBranch(in.Name, in.Ref)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	h.logic.RecordCreateBranchHint(in.Name)
	Ok(c, b)
}

func (h *Handler) CreateMergeRequest(c *app.RequestContext) {
	var in struct {
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		Squash       bool   `json:"squash"`
		RemoveSource bool   `json:"remove_source_branch"`
		MWPS         bool   `json:"merge_when_pipeline_succeeds"`
	}
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	if in.SourceBranch == "" || in.TargetBranch == "" || in.Title == "" {
		Err(c, 400, "source_branch/target_branch/title required")
		return
	}
	if in.SourceBranch == in.TargetBranch {
		Err(c, 400, "source and target branch must differ")
		return
	}
	_, _, mr, err := h.logic.CreateMergeRequest(gitlab.CreateMRInput{
		SourceBranch: in.SourceBranch,
		TargetBranch: in.TargetBranch,
		Title:        in.Title,
		Description:  in.Description,
		Squash:       in.Squash,
		RemoveSource: in.RemoveSource,
		MWPS:         in.MWPS,
	})
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	if mr != nil {
		h.logic.RecordMergeBranchHintWithURL(in.TargetBranch, mr.WebURL, in.Title)
	}
	Ok(c, mr)
}

func (h *Handler) AcceptMergeRequest(c *app.RequestContext) {
	idStr := string(c.Param("iid"))
	iid, err := strconv.Atoi(idStr)
	if err != nil {
		Err(c, 400, "invalid iid")
		return
	}
	var in struct {
		Squash       bool   `json:"squash"`
		RemoveSource bool   `json:"remove_source_branch"`
		MWPS         bool   `json:"merge_when_pipeline_succeeds"`
		Message      string `json:"merge_commit_message"`
	}
	_ = c.Bind(&in)
	mr, err := h.logic.AcceptMergeRequest(iid, in.Squash, in.RemoveSource, in.MWPS, in.Message)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	if mr != nil {
		h.logic.RecordMergeBranchHint(mr.TargetBranch)
	}
	Ok(c, mr)
}

func (h *Handler) AutoMerge(c *app.RequestContext) {
	var in struct {
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		Squash       bool   `json:"squash"`
		RemoveSource bool   `json:"remove_source_branch"`
		MWPS         bool   `json:"merge_when_pipeline_succeeds"`
		Message      string `json:"merge_commit_message"`
	}
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	if in.SourceBranch == "" || in.TargetBranch == "" || in.Title == "" {
		Err(c, 400, "source_branch/target_branch/title required")
		return
	}
	if in.SourceBranch == in.TargetBranch {
		Err(c, 400, "source and target branch must differ")
		return
	}
	_, _, mr, err := h.logic.CreateMergeRequest(gitlab.CreateMRInput{
		SourceBranch: in.SourceBranch,
		TargetBranch: in.TargetBranch,
		Title:        in.Title,
		Description:  in.Description,
		Squash:       in.Squash,
		RemoveSource: in.RemoveSource,
		MWPS:         in.MWPS,
	})
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	if mr == nil {
		Err(c, 500, "failed to create or find merge request")
		return
	}
	acc, err := h.logic.AcceptMergeRequest(mr.IID, in.Squash, in.RemoveSource, in.MWPS, in.Message)
	if err != nil {
		Err(c, 500, "auto merge failed, please merge manually: "+err.Error())
		return
	}
	Ok(c, acc)
}

func (h *Handler) Promote(c *app.RequestContext) {
	var in struct {
		SourcePrefix string `json:"source_prefix"`
		Name         string `json:"name"`
		Target       string `json:"target"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		Squash       bool   `json:"squash"`
		RemoveSource bool   `json:"remove_source_branch"`
		MWPS         bool   `json:"merge_when_pipeline_succeeds"`
		Message      string `json:"merge_commit_message"`
	}
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	if in.SourcePrefix == "" || in.Name == "" {
		Err(c, 400, "source_prefix/name required")
		return
	}
	mr, err := h.logic.Promote(gitlab.PromoteInput{
		SourcePrefix: in.SourcePrefix,
		Name:         in.Name,
		Target:       in.Target,
		Title:        in.Title,
		Description:  in.Description,
		Squash:       in.Squash,
		RemoveSource: in.RemoveSource,
		MWPS:         in.MWPS,
	})
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	if mr == nil {
		Ok(c, nil)
		return
	}
	acc, err := h.logic.AcceptMergeRequest(mr.IID, in.Squash, in.RemoveSource, in.MWPS, in.Message)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, acc)
}
