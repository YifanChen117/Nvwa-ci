package branch

import (
	"time"
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gorm.io/gorm"
)

// Service 分支服务
// 封装分支的业务逻辑与 go-git 集成
type Service struct {
    db       *gorm.DB
    repoPath string
    branches *repository.BranchRepository
}

// NewService 创建分支服务实例
func NewService(db *gorm.DB, repoPath string) *Service {
    return &Service{db: db, repoPath: repoPath, branches: repository.NewBranchRepository(db)}
}

func (s *Service) SetRepoPath(path string) { s.repoPath = path }

// RefreshBranch 从本地仓库刷新分支信息
// 根据分支名称查询最新 commit 与作者写入模型
func (s *Service) RefreshBranch(b *model.Branch) error {
	// 未配置仓库路径则跳过（演示模式）
	if s.repoPath == "" {
		return nil
	}
	r, err := git.PlainOpen(s.repoPath)
	if err != nil {
		return err
	}
	refName := plumbing.ReferenceName("refs/heads/" + b.Name)
	ref, err := r.Reference(refName, true)
	if err != nil {
		return err
	}
	commitHash := ref.Hash()
	commitObj, err := r.CommitObject(commitHash)
	if err != nil {
		return err
	}
	// 将最新提交信息回写到模型，便于页面展示
	b.LastCommitID = commitHash.String()
	b.LastPushTime = commitObj.Author.When
	b.LastPusher = commitObj.Author.Name
	b.UpdatedAt = time.Now()
	return s.branches.Update(b)
}
