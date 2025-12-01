package dal

import (
	"webci-refactored/internal/dal/model"

	"gorm.io/gorm"
)

// AutoMigrate 执行模型自动迁移
// 迁移 branches、environments、jobs 三张表结构
func AutoMigrate(db *gorm.DB) error {
	// GORM 根据结构体与标签生成/更新表结构，保证开发与数据库一致
	return db.AutoMigrate(&model.Branch{}, &model.Environment{}, &model.Job{})
}
