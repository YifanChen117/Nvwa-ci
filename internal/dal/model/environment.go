package model

import (
	"time"

	"gorm.io/gorm"
)

// Environment 环境模型
// 映射 environments 表，存储环境名称与描述
type Environment struct {
	// 主键：自增 ID
	ID uint64 `gorm:"primaryKey" json:"id"`
	// 名称：唯一索引，区分不同部署环境（prod/staging 等）
	Name string `gorm:"size:128;uniqueIndex" json:"name"`
	// 描述：环境说明或用途
	Description string `gorm:"size:255" json:"description"`
	// 当前部署信息
	CurrentDeployCommit string    `gorm:"size:64" json:"current_deploy_commit"`
	CurrentDeployAt     time.Time `gorm:"type:datetime" json:"current_deploy_at"`
	// 审计时间戳
	CreatedAt time.Time `gorm:"type:datetime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime" json:"updated_at"`
	// 软删除标记
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 返回表名
func (Environment) TableName() string { return "environments" }
