package repository

import (
	"webci-refactored/internal/dal/model"

	"gorm.io/gorm"
)

// EnvironmentRepository 环境仓库
// 提供环境的 CRUD 与查询方法
type EnvironmentRepository struct{ db *gorm.DB }

// NewEnvironmentRepository 创建环境仓库实例
func NewEnvironmentRepository(db *gorm.DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

// List 分页列出环境
func (r *EnvironmentRepository) List(limit, offset int) ([]model.Environment, int64, error) {
	// 两段式分页：先 Count 总数，再按更新时间倒序分页查数据
	var items []model.Environment
	var total int64
	if err := r.db.Model(&model.Environment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Get 获取环境
func (r *EnvironmentRepository) Get(id uint64) (*model.Environment, error) {
	var e model.Environment
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// Create 创建环境
func (r *EnvironmentRepository) Create(e *model.Environment) error { return r.db.Create(e).Error }

// Update 更新环境
func (r *EnvironmentRepository) Update(e *model.Environment) error { return r.db.Save(e).Error }

// Delete 删除环境
func (r *EnvironmentRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Environment{}, id).Error
}

// GetByName 通过名称获取环境
func (r *EnvironmentRepository) GetByName(name string) (*model.Environment, error) {
	// 幂等：若名称存在则直接返回现有记录，便于简化前端/脚本处理
	var e model.Environment
	if err := r.db.Where("name = ?", name).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}
