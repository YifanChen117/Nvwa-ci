package job

import (
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"
	"webci-refactored/internal/queue"
	"webci-refactored/internal/service/job"

	"gorm.io/gorm"
)

// Logic 任务业务逻辑
type Logic struct {
	db   *gorm.DB
	svc  *job.Service
	jobs *repository.JobRepository
}

// NewLogic 创建任务业务逻辑实例
func NewLogic(db *gorm.DB) *Logic {
	return &Logic{
		db:   db,
		svc:  job.NewService(db),
		jobs: repository.NewJobRepository(db),
	}
}

// Create 创建任务
func (l *Logic) Create(branchID, envID uint64, triggerUser string) (*model.Job, error) {
	// 绑定请求体并校验必要字段（服务层会对 trigger_user 做二次校验）
	j, err := l.svc.Create(branchID, envID, triggerUser)
	if err != nil {
		return nil, err
	}
	// 推入队列，异步执行器开始处理
	queue.Enqueue(j.ID)
	return j, nil
}

// List 查询任务
func (l *Logic) List(branchID, envID *uint64, status *string, limit, offset int) ([]model.Job, int64, error) {
	return l.jobs.List(branchID, envID, status, limit, offset)
}

// Get 获取任务详情
func (l *Logic) Get(id uint64) (*model.Job, error) {
	return l.jobs.Get(id)
}

// UpdateStatus 更新任务状态
func (l *Logic) UpdateStatus(id uint64, status, logAppend string) (*model.Job, error) {
	// 状态与日志均为可选：只要有传入就执行对应更新
	if status != "" {
		if err := l.svc.SetStatus(id, status); err != nil {
			return nil, err
		}
	}
	if logAppend != "" {
		if err := l.svc.AppendLog(id, logAppend); err != nil {
			return nil, err
		}
	}
	return l.jobs.Get(id)
}

// Log 单独获取构建日志
func (l *Logic) Log(id uint64) (string, error) {
	j, err := l.jobs.Get(id)
	if err != nil {
		return "", err
	}
	// 直接返回纯文本日志
	return j.Log, nil
}

// Cancel 取消任务
func (l *Logic) Cancel(id uint64) error {
	// 标记取消供执行器识别，并追加取消请求的日志
	queue.MarkCancel(id)
	if err := l.svc.Cancel(id); err != nil {
		return err
	}
	return nil
}
