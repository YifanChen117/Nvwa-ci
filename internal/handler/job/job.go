package job

import (
	"strconv"
	"webci-refactored/internal/logic/job"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

// Handler 任务处理层
type Handler struct {
	logic *job.Logic
}

// NewHandler 创建任务处理层实例
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		logic: job.NewLogic(db),
	}
}

// Ok 返回成功响应
func Ok(c *app.RequestContext, data interface{}) {
	c.JSON(200, map[string]interface{}{"code": 0, "message": "ok", "data": data})
}

// Err 返回错误响应
func Err(c *app.RequestContext, status int, msg string) {
	c.JSON(status, map[string]interface{}{"code": status, "message": msg})
}

// parseID 从路径参数中解析ID
func parseID(c *app.RequestContext) uint64 {
	idStr := string(c.Param("id"))
	id, _ := strconv.ParseUint(idStr, 10, 64)
	return id
}

// Create 创建任务
func (h *Handler) Create(c *app.RequestContext) {
	var in struct {
		BranchID    uint64 `json:"branch_id"`
		EnvID       uint64 `json:"env_id"`
		TriggerUser string `json:"trigger_user"`
	}
	// 绑定请求体并校验必要字段
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	j, err := h.logic.Create(in.BranchID, in.EnvID, in.TriggerUser)
	if err != nil {
		Err(c, 400, err.Error())
		return
	}
	Ok(c, j)
}

// List 查询任务
func (h *Handler) List(c *app.RequestContext) {
	var limit, offset = 50, 0
	var branchIDPtr, envIDPtr *uint64
	var statusPtr *string
	// Query 解析：有值则构建指针供逻辑层选择性过滤
	if v := c.Query("branch_id"); len(v) > 0 {
		id, _ := strconv.ParseUint(string(v), 10, 64)
		branchIDPtr = &id
	}
	if v := c.Query("env_id"); len(v) > 0 {
		id, _ := strconv.ParseUint(string(v), 10, 64)
		envIDPtr = &id
	}
	if v := c.Query("status"); len(v) > 0 {
		s := string(v)
		statusPtr = &s
	}
	items, total, err := h.logic.List(branchIDPtr, envIDPtr, statusPtr, limit, offset)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, map[string]interface{}{"items": items, "total": total})
}

// Get 获取任务详情
func (h *Handler) Get(c *app.RequestContext) {
	id := parseID(c)
	j, err := h.logic.Get(id)
	if err != nil {
		Err(c, 404, err.Error())
		return
	}
	Ok(c, j)
}

// UpdateStatus 更新任务状态
func (h *Handler) UpdateStatus(c *app.RequestContext) {
	id := parseID(c)
	var in struct {
		Status    string `json:"status"`
		LogAppend string `json:"log_append"`
	}
	// 状态与日志均为可选：只要有传入就执行对应更新
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	j, err := h.logic.UpdateStatus(id, in.Status, in.LogAppend)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, j)
}

// Log 单独获取构建日志
func (h *Handler) Log(c *app.RequestContext) {
	id := parseID(c)
	log, err := h.logic.Log(id)
	if err != nil {
		Err(c, 404, err.Error())
		return
	}
	// 直接返回纯文本日志，适配前端的 log 视图
	c.String(200, log)
}

// Cancel 取消任务
func (h *Handler) Cancel(c *app.RequestContext) {
	id := parseID(c)
	// 标记取消供执行器识别
	if err := h.logic.Cancel(id); err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, "cancelled")
}
