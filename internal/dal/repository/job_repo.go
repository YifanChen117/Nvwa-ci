package repository

import (
	"time"
	"webci-refactored/internal/dal/model"

	"gorm.io/gorm"
)

// JobRepository 任务仓库
// 提供任务的创建、查询、状态更新与日志追加
type JobRepository struct{ db *gorm.DB }

// NewJobRepository 创建任务仓库实例
func NewJobRepository(db *gorm.DB) *JobRepository { return &JobRepository{db: db} }

// Create 创建任务
func (r *JobRepository) Create(j *model.Job) error { return r.db.Create(j).Error }

// Get 获取任务详情
func (r *JobRepository) Get(id uint64) (*model.Job, error) {
	var j model.Job
	if err := r.db.First(&j, id).Error; err != nil {
		return nil, err
	}
	return &j, nil
}

// List 条件查询任务
func (r *JobRepository) List(branchID, envID *uint64, status *string, limit, offset int) ([]model.Job, int64, error) {
	// 可选过滤：当指针非空时添加 Where 子句；按 id 倒序便于查看最新任务
	q := r.db.Model(&model.Job{})
	if branchID != nil {
		q = q.Where("branch_id = ?", *branchID)
	}
	if envID != nil {
		q = q.Where("env_id = ?", *envID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Job
	if err := q.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateStatus 更新任务状态
func (r *JobRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).Update("status", status).Error
}

// AppendLog 追加任务日志
func (r *JobRepository) AppendLog(id uint64, text string) error {
	// 兼容不同方言的字符串追加：MySQL 使用 CONCAT；SQLite 使用 ||
	expr := gorm.Expr("CONCAT(IFNULL(log,''), ?)", text)
	if r.db.Dialector.Name() == "sqlite" {
		expr = gorm.Expr("COALESCE(log,'') || ?", text)
	}
	return r.db.Model(&model.Job{}).Where("id = ?", id).Update("log", expr).Error
}

// UpdateTimes 更新开始/结束时间
func (r *JobRepository) UpdateTimes(id uint64, start, end *time.Time) error {
	// 仅更新非空的时间字段，避免覆盖已有值
	data := map[string]interface{}{}
	if start != nil {
		data["start_time"] = *start
	}
	if end != nil {
		data["end_time"] = *end
	}
	return r.db.Model(&model.Job{}).Where("id = ?", id).Updates(data).Error
}
