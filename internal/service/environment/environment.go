package environment

import (
	"errors"
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"

	"gorm.io/gorm"
)

// Service 环境服务
// 提供环境的校验与业务封装
type Service struct {
	db   *gorm.DB
	envs *repository.EnvironmentRepository
}

// NewService 创建环境服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db, envs: repository.NewEnvironmentRepository(db)}
}

// Create 创建环境，校验名称非空与唯一
func (s *Service) Create(e *model.Environment) error {
	// 名称不能为空
	if e.Name == "" {
		return errors.New("environment name required")
	}
	var count int64
	if err := s.db.Model(&model.Environment{}).Where("name = ?", e.Name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		// 幂等：名称已存在时返回现有记录（不报错），减少前端/脚本复杂度
		exist, err := s.envs.GetByName(e.Name)
		if err != nil {
			return err
		}
		e.ID = exist.ID
		e.Description = exist.Description
		return nil
	}
	return s.envs.Create(e)
}
