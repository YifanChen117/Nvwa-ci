package dashboard

import (
	"webci-refactored/internal/service/dashboard"

	"gorm.io/gorm"
)

// Logic 仪表盘业务逻辑
type Logic struct {
	svc *dashboard.Service
}

// NewLogic 创建仪表盘业务逻辑实例
func NewLogic(db *gorm.DB) *Logic {
	return &Logic{
		svc: dashboard.NewService(db),
	}
}

// Overview 总览统计
func (l *Logic) Overview() (map[string]int64, error) {
	// 返回各状态的数量与总数，便于页面总览展示
	return l.svc.Overview()
}

// BranchStats 分支维度统计
func (l *Logic) BranchStats(branchID uint64) (map[string]int64, error) {
	// 返回该分支的任务状态分布
	return l.svc.BranchStats(branchID)
}

// EnvironmentStats 环境维度统计
func (l *Logic) EnvironmentStats(envID uint64) (map[string]int64, error) {
	// 返回该环境的任务状态分布
	return l.svc.EnvironmentStats(envID)
}
