package environment

import (
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"
	"webci-refactored/internal/service/environment"

	"gorm.io/gorm"
)

// Logic 环境业务逻辑
type Logic struct {
	db   *gorm.DB
	svc  *environment.Service
	envs *repository.EnvironmentRepository
}

// NewLogic 创建环境业务逻辑实例
func NewLogic(db *gorm.DB) *Logic {
	return &Logic{
		db:   db,
		svc:  environment.NewService(db),
		envs: repository.NewEnvironmentRepository(db),
	}
}

// List 列出环境
func (l *Logic) List(limit, offset int) ([]model.Environment, int64, error) {
	return l.envs.List(limit, offset)
}

// Create 创建环境
func (l *Logic) Create(name, description string) (*model.Environment, error) {
	e := &model.Environment{Name: name, Description: description}
	// 服务层校验名称非空与唯一；若存在返回现有记录实现幂等
	if err := l.svc.Create(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Get 获取环境
func (l *Logic) Get(id uint64) (*model.Environment, error) {
	return l.envs.Get(id)
}

// Update 更新环境
func (l *Logic) Update(id uint64, name, description string) (*model.Environment, error) {
	e, err := l.envs.Get(id)
	if err != nil {
		return nil, err
	}
	// 更新基本字段并持久化
	e.Name = name
	e.Description = description
	if err := l.envs.Update(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Delete 删除环境
func (l *Logic) Delete(id uint64) error {
	return l.envs.Delete(id)
}
