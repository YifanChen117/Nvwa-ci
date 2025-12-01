package repository

import (
	"webci-refactored/internal/dal/model"

	"gorm.io/gorm"
)

// BranchRepository 分支仓库
// 提供分支的 CRUD 与查询方法
type BranchRepository struct {
	db *gorm.DB
}

// NewBranchRepository 创建分支仓库实例
func NewBranchRepository(db *gorm.DB) *BranchRepository { return &BranchRepository{db: db} }

// List 分页列出分支
func (r *BranchRepository) List(limit, offset int) ([]model.Branch, int64, error) {
	// 先统计总数，再按更新时间倒序分页查询，便于看到最近更新的分支
	var items []model.Branch
	var total int64
	if err := r.db.Model(&model.Branch{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Get 获取分支
func (r *BranchRepository) Get(id uint64) (*model.Branch, error) {
	var b model.Branch
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// Create 创建分支
func (r *BranchRepository) Create(b *model.Branch) error { return r.db.Create(b).Error }

// Update 更新分支
func (r *BranchRepository) Update(b *model.Branch) error { return r.db.Save(b).Error }

// Delete 删除分支
func (r *BranchRepository) Delete(id uint64) error { return r.db.Delete(&model.Branch{}, id).Error }

// GetByName 通过名称获取分支
func (r *BranchRepository) GetByName(name string) (*model.Branch, error) {
	// 幂等：当名称已存在时返回现有记录而不是报错，便于脚本与前端简化处理
	var b model.Branch
	if err := r.db.Where("name = ?", name).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}
