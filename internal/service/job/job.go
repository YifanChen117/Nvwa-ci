package job

import (
	"errors"
	"fmt"
	"time"
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"

	"gorm.io/gorm"
)

// Service 任务服务
// 封装任务创建、状态机与日志操作
type Service struct {
	db   *gorm.DB
	jobs *repository.JobRepository
}

// NewService 创建任务服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db, jobs: repository.NewJobRepository(db)}
}

// Create 创建 pending 任务并推入队列
func (s *Service) Create(branchID, envID uint64, triggerUser string) (*model.Job, error) {
	// 基础校验：触发用户必填
	if triggerUser == "" {
		return nil, errors.New("trigger_user required")
	}
	// 构造初始任务：状态 pending，等待执行器接手
	j := &model.Job{BranchID: branchID, EnvID: envID, Status: "pending", TriggerUser: triggerUser}
	if err := s.jobs.Create(j); err != nil {
		return nil, err
	}
	return j, nil
}

// AppendLog 追加日志
func (s *Service) AppendLog(id uint64, text string) error { return s.jobs.AppendLog(id, text) }

// SetStatus 设置任务状态
func (s *Service) SetStatus(id uint64, status string) error {
	// 直接透传到仓库层更新状态
	return s.jobs.UpdateStatus(id, status)
}

// Cancel 标记任务取消
func (s *Service) Cancel(id uint64) error {
	// 仅追加一条取消请求的日志；真正的取消由执行器在运行阶段识别
	return s.jobs.AppendLog(id, fmt.Sprintf("\n[INFO] cancel requested at %s", time.Now().Format(time.RFC3339)))
}
