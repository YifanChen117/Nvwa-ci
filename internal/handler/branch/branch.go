package branch

import (
	"strconv"
	"webci-refactored/internal/logic/branch"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

// Handler 分支处理层
type Handler struct {
    logic *branch.Logic
}

// NewHandler 创建分支处理层实例
func NewHandler(db *gorm.DB, repoPath string) *Handler {
	return &Handler{
		logic: branch.NewLogic(db, repoPath),
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

// List 列出分支
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

// Create 创建分支
func (h *Handler) Create(c *app.RequestContext) {
	var in struct{ Name string }
	// 解析 JSON 请求体，绑定到匿名结构体
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	b, err := h.logic.Create(in.Name)
	if err != nil {
		Err(c, 400, err.Error())
		return
	}
	Ok(c, b)
}

// Get 获取分支
func (h *Handler) Get(c *app.RequestContext) {
	// 从路径参数读取 id，并转换为无符号整数
	id := parseID(c)
	b, err := h.logic.Get(id)
	if err != nil {
		Err(c, 404, err.Error())
		return
	}
	Ok(c, b)
}

// Update 更新分支
func (h *Handler) Update(c *app.RequestContext) {
	id := parseID(c)
	var in struct{ Name string }
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	b, err := h.logic.Update(id, in.Name)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, b)
}

// Delete 删除分支
func (h *Handler) Delete(c *app.RequestContext) {
	id := parseID(c)
	if err := h.logic.Delete(id); err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, "deleted")
}

// Refresh 刷新分支信息
func (h *Handler) Refresh(c *app.RequestContext) {
	id := parseID(c)
	b, err := h.logic.Refresh(id)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, b)
}

// MockPush 模拟推送
func (h *Handler) MockPush(c *app.RequestContext) {
	id := parseID(c)
	var in struct {
		Pusher      string  `json:"pusher"`
		EnvID       *uint64 `json:"env_id"`
		TriggerUser string  `json:"trigger_user"`
		Message     string  `json:"message"`
	}
	if err := c.Bind(&in); err != nil {
		Err(c, 400, err.Error())
		return
	}
	b, job, err := h.logic.MockPush(id, in.Pusher, in.EnvID, in.TriggerUser, in.Message)
	if err != nil {
		Err(c, 500, err.Error())
		return
	}
	Ok(c, map[string]interface{}{"branch": b, "job": job})
}

func (h *Handler) UpdateRepoPath(path string) { h.logic.UpdateRepoPath(path) }
