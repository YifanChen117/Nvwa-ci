package branch

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"
	"webci-refactored/internal/queue"
	"webci-refactored/internal/service/branch"
	"webci-refactored/internal/service/job"

	"gorm.io/gorm"
)

// Logic 分支业务逻辑
type Logic struct {
	db       *gorm.DB
	svc      *branch.Service
	jobSvc   *job.Service
	branches *repository.BranchRepository
}

// NewLogic 创建分支业务逻辑实例
func NewLogic(db *gorm.DB, repoPath string) *Logic {
	return &Logic{
		db:       db,
		svc:      branch.NewService(db, repoPath),
		jobSvc:   job.NewService(db),
		branches: repository.NewBranchRepository(db),
	}
}

func (l *Logic) UpdateRepoPath(path string) { l.svc.SetRepoPath(path) }

// List 列出分支
func (l *Logic) List(limit, offset int) ([]model.Branch, int64, error) {
	return l.branches.List(limit, offset)
}

// Create 创建分支
func (l *Logic) Create(name string) (*model.Branch, error) {
	if name == "" {
		return nil, &Error{"name required"}
	}
	// 幂等：若名称已存在，直接返回现有记录
	if existing, err := l.branches.GetByName(name); err == nil {
		return existing, nil
	}
	b := &model.Branch{Name: name, UpdatedAt: time.Now()}
	if err := l.branches.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Get 获取分支
func (l *Logic) Get(id uint64) (*model.Branch, error) {
	return l.branches.Get(id)
}

// Update 更新分支
func (l *Logic) Update(id uint64, name string) (*model.Branch, error) {
	b, err := l.branches.Get(id)
	if err != nil {
		return nil, err
	}
	// 更新名称与更新时间
	b.Name = name
	b.UpdatedAt = time.Now()
	if err := l.branches.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Delete 删除分支
func (l *Logic) Delete(id uint64) error {
	return l.branches.Delete(id)
}

// Refresh 刷新分支信息
func (l *Logic) Refresh(id uint64) (*model.Branch, error) {
	b, err := l.branches.Get(id)
	if err != nil {
		return nil, err
	}
	// 调用服务层集成 go-git 刷新最近提交信息
	if err := l.svc.RefreshBranch(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MockPush 模拟一次用户推送：生成随机 commit，并可选触发 CI 任务
func (l *Logic) MockPush(id uint64, pusher string, envID *uint64, triggerUser, message string) (*model.Branch, *model.Job, error) {
	b, err := l.branches.Get(id)
	if err != nil {
		return nil, nil, err
	}
	if pusher == "" {
		pusher = "demo"
	}
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return nil, nil, err
	}
	hash := hex.EncodeToString(buf)
	now := time.Now()
	b.LastCommitID = hash
	b.LastPushTime = now
	b.LastPusher = pusher
	b.UpdatedAt = now
	if err := l.branches.Update(b); err != nil {
		return nil, nil, err
	}
	var job *model.Job
	if envID != nil {
		user := triggerUser
		if user == "" {
			user = pusher
		}
		j, err := l.jobSvc.Create(id, *envID, user)
		if err != nil {
			return nil, nil, err
		}
		queue.Enqueue(j.ID)
		if message != "" {
			_ = l.jobSvc.AppendLog(j.ID, "[COMMIT] "+hash+" "+message+"\n")
		} else {
			_ = l.jobSvc.AppendLog(j.ID, "[COMMIT] "+hash+"\n")
		}
		job = j
	}
	return b, job, nil
}

// Error 自定义错误类型
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
