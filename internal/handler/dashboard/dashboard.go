package dashboard

import (
	"strconv"
	"webci-refactored/internal/logic/dashboard"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

// Handler 仪表盘处理层
type Handler struct {
	logic *dashboard.Logic
}

// NewHandler 创建仪表盘处理层实例
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		logic: dashboard.NewLogic(db),
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

// Overview 总览统计
func (h *Handler) Overview(c *app.RequestContext) {
	// 返回各状态的数量与总数，便于页面总览展示
	data, err := h.logic.Overview()
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, data)
}

// BranchStats 分支维度统计
func (h *Handler) BranchStats(c *app.RequestContext) {
	id := parseID(c)
	// 返回该分支的任务状态分布
	data, err := h.logic.BranchStats(id)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, data)
}

// EnvironmentStats 环境维度统计
func (h *Handler) EnvironmentStats(c *app.RequestContext) {
	id := parseID(c)
	// 返回该环境的任务状态分布
	data, err := h.logic.EnvironmentStats(id)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, data)
}
