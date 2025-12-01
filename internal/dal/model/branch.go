package model

import (
	"time"

	"gorm.io/gorm"
)

// Branch 分支模型
// 映射 branches 表，存储最近提交信息与推送者等
type Branch struct {
	// 主键：自增 ID，对外 JSON 为 id
	ID uint64 `gorm:"primaryKey" json:"id"`
	// 名称：唯一索引，避免重复分支名
	Name string `gorm:"size:128;uniqueIndex" json:"name"`
	// 最近提交哈希与推送者信息
	LastCommitID string    `gorm:"size:64" json:"last_commit_id"`
	LastPushTime time.Time `gorm:"type:datetime" json:"last_push_time"`
	LastPusher   string    `gorm:"size:64" json:"last_pusher"`
	// 审计时间戳（创建/更新），以 datetime 存储
	CreatedAt time.Time `gorm:"type:datetime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:datetime" json:"updated_at"`
	// 软删除标记：支持安全删除与数据恢复
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 返回表名
func (Branch) TableName() string { return "branches" }
