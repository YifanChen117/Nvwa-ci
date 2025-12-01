package router

import (
	"context"
	"log"
	"webci-refactored/internal/config"
	"webci-refactored/internal/handler/branch"
	"webci-refactored/internal/handler/dashboard"
	"webci-refactored/internal/handler/environment"
	"webci-refactored/internal/handler/gitlab"
	"webci-refactored/internal/handler/job"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"gorm.io/gorm"
)

// indexPageHandler 首页处理器：直接显示GitLab流水线页面
func indexPageHandler(c context.Context, ctx *app.RequestContext) {
	// 如果GitLab处理器可用，直接显示GitLab页面
	// 否则显示简单的HTML页面提示用户访问GitLab页面
	cfg := config.Load()

	// 创建GitLab处理器
	gitlabHandler, err := gitlab.NewHandler(cfg)
	if err != nil {
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.SetStatusCode(200)
		ctx.Write([]byte(`<html><body><h1>WebCI</h1><p><a href="/gitlab">访问GitLab流水线页面</a></p></body></html>`))
		return
	}

	// GitLab处理器创建成功，显示GitLab页面内容
	html := gitlabHandler.PageContent()
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.SetStatusCode(200)
	ctx.Write([]byte(html))
}

// NewServer 创建并配置 Hertz 服务
// 注册路由、绑定中间件、挂载静态资源
func NewServer(cfg config.Config, db *gorm.DB) *server.Hertz {
	// 1) 初始化各模块处理器
	branchHandler := branch.NewHandler(db, cfg.RepoPath)
	envHandler := environment.NewHandler(db)
	jobHandler := job.NewHandler(db)
	dashboardHandler := dashboard.NewHandler(db)

	// 创建GitLab处理器
	log.Printf("Creating GitLab handler with config - BaseURL: %s, Project: %s", cfg.GitLabBaseURL, cfg.GitLabProject)
	gitlabHandler, err := gitlab.NewHandler(cfg)
	if err != nil {
		log.Printf("Warning: failed to create gitlab handler: %v", err)
		gitlabHandler = nil
	} else {
		log.Printf("Successfully created GitLab handler")
	}

	// 2) 创建 Hertz 服务实例
	h := server.Default(server.WithHostPorts(cfg.HTTPAddr))

	// 3) 注册 API 路由
	api := h.Group("/api")
	{
		// 分支相关路由
		branches := api.Group("/branches")
		{
			branches.GET("", branchListHandler(branchHandler))
			branches.POST("", branchCreateHandler(branchHandler))
			branches.GET("/:id", branchGetHandler(branchHandler))
			branches.PUT("/:id", branchUpdateHandler(branchHandler))
			branches.DELETE("/:id", branchDeleteHandler(branchHandler))
			branches.POST("/:id/refresh", branchRefreshHandler(branchHandler))
			branches.POST("/:id/mock_push", branchMockPushHandler(branchHandler))
		}

		// 环境相关路由
		envs := api.Group("/environments")
		{
			envs.GET("", envListHandler(envHandler))
			envs.POST("", envCreateHandler(envHandler))
			envs.GET("/:id", envGetHandler(envHandler))
			envs.PUT("/:id", envUpdateHandler(envHandler))
			envs.DELETE("/:id", envDeleteHandler(envHandler))
		}

		// 任务相关路由
		jobs := api.Group("/jobs")
		{
			jobs.GET("", jobListHandler(jobHandler))
			jobs.POST("", jobCreateHandler(jobHandler))
			jobs.GET("/:id", jobGetHandler(jobHandler))
			jobs.PUT("/:id/status", jobUpdateStatusHandler(jobHandler))
			jobs.GET("/:id/log", jobLogHandler(jobHandler))
			jobs.POST("/:id/cancel", jobCancelHandler(jobHandler))
		}

		// 仪表盘相关路由
		dashboardGroup := api.Group("/dashboard")
		{
			dashboardGroup.GET("/overview", dashboardOverviewHandler(dashboardHandler))
			dashboardGroup.GET("/branch/:id", dashboardBranchStatsHandler(dashboardHandler))
			dashboardGroup.GET("/environment/:id", dashboardEnvironmentStatsHandler(dashboardHandler))
		}

		gitlabAPI := api.Group("/gitlab")
		gitlabAPI.GET("/config", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(200, map[string]interface{}{
				"code":    0,
				"message": "ok",
				"data": map[string]string{
					"base_url":   cfg.GitLabBaseURL,
					"project_id": cfg.GitLabProject,
					"repo_path":  cfg.RepoPath,
				},
			})
		})
		gitlabAPI.POST("/config", func(c context.Context, ctx *app.RequestContext) {
			var in struct {
				BaseURL   string `json:"base_url"`
				Token     string `json:"token"`
				ProjectID string `json:"project_id"`
				RepoPath  string `json:"repo_path"`
			}
			if err := ctx.Bind(&in); err != nil {
				ctx.JSON(400, map[string]interface{}{"code": 400, "message": err.Error()})
				return
			}
			newCfg := cfg
			if in.BaseURL != "" {
				newCfg.GitLabBaseURL = in.BaseURL
			}
			if in.Token != "" {
				newCfg.GitLabToken = in.Token
			}
			if in.ProjectID != "" {
				newCfg.GitLabProject = in.ProjectID
			}
			if in.RepoPath != "" {
				newCfg.RepoPath = in.RepoPath
			}
			if gitlabHandler != nil {
				if err := gitlabHandler.UpdateConfig(newCfg); err != nil {
					ctx.JSON(400, map[string]interface{}{"code": 400, "message": err.Error()})
					return
				}
			}
			cfg = newCfg
			if in.RepoPath != "" {
				branchHandler.UpdateRepoPath(in.RepoPath)
			}
			ctx.JSON(200, map[string]interface{}{
				"code":    0,
				"message": "ok",
				"data": map[string]string{
					"base_url":   cfg.GitLabBaseURL,
					"project_id": cfg.GitLabProject,
					"repo_path":  cfg.RepoPath,
				},
			})
		})
		if gitlabHandler != nil {
			gitlabAPI.GET("/pipelines", gitlabListPipelinesHandler(gitlabHandler))
			gitlabAPI.GET("/pipelines/:id", gitlabGetPipelineHandler(gitlabHandler))
			gitlabAPI.GET("/branches", gitlabListBranchesHandler(gitlabHandler))
			gitlabAPI.GET("/jobs", gitlabListJobsHandler(gitlabHandler))
			gitlabAPI.POST("/branches", gitlabCreateBranchHandler(gitlabHandler))
			gitlabAPI.POST("/merge_requests", gitlabCreateMRHandler(gitlabHandler))
			gitlabAPI.POST("/merge_requests/:iid/merge", gitlabAcceptMRHandler(gitlabHandler))
			gitlabAPI.POST("/promote", gitlabPromoteHandler(gitlabHandler))
			gitlabAPI.POST("/merge_requests/auto", gitlabAutoMergeHandler(gitlabHandler))
		}
	}

	// 4) 注册首页路由
	h.GET("/", indexPageHandler)

	// 5) 注册GitLab页面路由（仅在GitLab处理器创建成功时注册）
	if gitlabHandler != nil {
		h.GET("/gitlab", gitlabPageHandler(gitlabHandler))
	}

	// 6) 注册兜底处理器：处理未匹配的路由，返回 index.html 实现 SPA 支持
	h.NoRoute(func(c context.Context, ctx *app.RequestContext) {
		// 对于 /api 前缀的请求，返回 404 JSON 响应
		if len(ctx.Request.URI().Path()) >= 4 && string(ctx.Request.URI().Path())[:4] == "/api" {
			ctx.JSON(404, map[string]interface{}{"code": 404, "message": "api not found"})
			return
		}
		// 其他请求返回首页
		indexPageHandler(c, ctx)
	})

	return h
}

// 分支处理器包装函数
func branchListHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.List(ctx) }
}

func branchCreateHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Create(ctx) }
}

func branchGetHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Get(ctx) }
}

func branchUpdateHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Update(ctx) }
}

func branchDeleteHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Delete(ctx) }
}

func branchRefreshHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Refresh(ctx) }
}

func branchMockPushHandler(h *branch.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.MockPush(ctx) }
}

// 环境处理器包装函数
func envListHandler(h *environment.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.List(ctx) }
}

func envCreateHandler(h *environment.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Create(ctx) }
}

func envGetHandler(h *environment.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Get(ctx) }
}

func envUpdateHandler(h *environment.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Update(ctx) }
}

func envDeleteHandler(h *environment.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Delete(ctx) }
}

// 任务处理器包装函数
func jobListHandler(h *job.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.List(ctx) }
}

func jobCreateHandler(h *job.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Create(ctx) }
}

func jobGetHandler(h *job.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Get(ctx) }
}

func jobUpdateStatusHandler(h *job.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.UpdateStatus(ctx) }
}

func jobLogHandler(h *job.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Log(ctx) }
}

func jobCancelHandler(h *job.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Cancel(ctx) }
}

// 仪表盘处理器包装函数
func dashboardOverviewHandler(h *dashboard.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Overview(ctx) }
}

func dashboardBranchStatsHandler(h *dashboard.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.BranchStats(ctx) }
}

func dashboardEnvironmentStatsHandler(h *dashboard.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.EnvironmentStats(ctx) }
}

// GitLab处理器包装函数
func gitlabListPipelinesHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.ListPipelines(ctx) }
}

func gitlabGetPipelineHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.GetPipeline(ctx) }
}

func gitlabListBranchesHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.ListBranches(ctx) }
}

// 新增的GitLab任务列表处理器包装函数
func gitlabListJobsHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.ListJobs(ctx) }
}

// GitLab页面处理器包装函数
func gitlabPageHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Page(ctx) }
}

func gitlabCreateBranchHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.CreateBranch(ctx) }
}

func gitlabCreateMRHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.CreateMergeRequest(ctx) }
}

func gitlabAcceptMRHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.AcceptMergeRequest(ctx) }
}

func gitlabPromoteHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.Promote(ctx) }
}

func gitlabAutoMergeHandler(h *gitlab.Handler) func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) { h.AutoMerge(ctx) }
}
