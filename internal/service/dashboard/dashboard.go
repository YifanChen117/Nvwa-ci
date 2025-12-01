package dashboard

import (
	"webci-refactored/internal/dal/model"

	"gorm.io/gorm"
)

// Service 仪表盘服务
// 提供汇总统计与分支/环境维度统计
type Service struct{ db *gorm.DB }

// NewService 创建仪表盘服务
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Overview 返回任务状态汇总
func (s *Service) Overview() (map[string]int64, error) {
	// 统计各状态数量，并附加总数
	res := map[string]int64{}
	statuses := []string{"pending", "running", "success", "failed"}
	for _, st := range statuses {
		var c int64
		if err := s.db.Model(&model.Job{}).Where("status = ?", st).Count(&c).Error; err != nil {
			return nil, err
		}
		res[st] = c
	}
	var total int64
	if err := s.db.Model(&model.Job{}).Count(&total).Error; err != nil {
		return nil, err
	}
	res["total"] = total
	return res, nil
}

// BranchStats 返回某分支任务汇总
func (s *Service) BranchStats(branchID uint64) (map[string]int64, error) {
	// 指定分支 ID 的状态分布
	res := map[string]int64{}
	statuses := []string{"pending", "running", "success", "failed"}
	for _, st := range statuses {
		var c int64
		if err := s.db.Model(&model.Job{}).Where("branch_id = ? AND status = ?", branchID, st).Count(&c).Error; err != nil {
			return nil, err
		}
		res[st] = c
	}
	var total int64
	if err := s.db.Model(&model.Job{}).Where("branch_id = ?", branchID).Count(&total).Error; err != nil {
		return nil, err
	}
	res["total"] = total
	return res, nil
}

// EnvironmentStats 返回某环境任务汇总
func (s *Service) EnvironmentStats(envID uint64) (map[string]int64, error) {
	// 指定环境 ID 的状态分布
	res := map[string]int64{}
	statuses := []string{"pending", "running", "success", "failed"}
	for _, st := range statuses {
		var c int64
		if err := s.db.Model(&model.Job{}).Where("env_id = ? AND status = ?", envID, st).Count(&c).Error; err != nil {
			return nil, err
		}
		res[st] = c
	}
	var total int64
	if err := s.db.Model(&model.Job{}).Where("env_id = ?", envID).Count(&total).Error; err != nil {
		return nil, err
	}
	res["total"] = total
	return res, nil
}
