package environment

import (
	"strconv"
	"webci-refactored/internal/logic/environment"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

// Handler 环境处理层
type Handler struct {
	logic *environment.Logic
}

// NewHandler 创建环境处理层实例
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		logic: environment.NewLogic(db),
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

// List 列出环境
func (h *Handler) List(c *app.RequestContext) {
	// 统一分页策略：固定 limit/offset，生产可改为请求参数
	limit, offset := 50, 0
	items, total, err := h.logic.List(limit, offset)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, map[string]interface{}{"items": items, "total": total})
}

// Create 创建环境
func (h *Handler) Create(c *app.RequestContext) {
	var in struct{ Name, Description string }
	// 绑定 JSON 请求体
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	e, err := h.logic.Create(in.Name, in.Description)
	if err != nil {
		Err(c, 400, err.Error())
		return
	}
	Ok(c, e)
}

// Get 获取环境
func (h *Handler) Get(c *app.RequestContext) {
	// 解析路径参数中的环境 ID
	id := parseID(c)
	e, err := h.logic.Get(id)
	if err != nil {
		Err(c, 404, err.Error())
		return
	}
	Ok(c, e)
}

// Update 更新环境
func (h *Handler) Update(c *app.RequestContext) {
	id := parseID(c)
	e, err := h.logic.Get(id)
	if err != nil {
		Err(c, 404, err.Error())
		return
	}
	var in struct{ Name, Description string }
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	// 更新基本字段并持久化
	e, err = h.logic.Update(id, in.Name, in.Description)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, e)
}

// Delete 删除环境
func (h *Handler) Delete(c *app.RequestContext) {
	id := parseID(c)
	if err := h.logic.Delete(id); err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, "deleted")
}
