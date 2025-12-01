package model

import (
	"time"

	"gorm.io/gorm"
)

// Job 构建任务模型
// 映射 jobs 表，包含状态机字段与日志
type Job struct {
	// 主键：任务 ID
	ID uint64 `gorm:"primaryKey" json:"id"`
	// 归属维度：分支与环境 ID，便于过滤查询
	BranchID uint64 `gorm:"index" json:"branch_id"`
	EnvID    uint64 `gorm:"index" json:"env_id"`
	// 状态机：pending → running → success/failed（index 便于统计与过滤）
	Status string `gorm:"size:16;default:'pending';index" json:"status"`
	// 触发用户：记录是谁发起任务
	TriggerUser string `gorm:"size:64" json:"trigger_user"`
	// 时间戳：开始/结束时间，便于度量耗时
	StartTime *time.Time `gorm:"type:datetime" json:"start_time"`
	EndTime   *time.Time `gorm:"type:datetime" json:"end_time"`
	// 日志：构建过程的文本输出
	Log string `gorm:"type:text" json:"log"`
	// Git提交信息
	CommitID      string     `gorm:"size:64" json:"commit_id"`
	CommitMessage string     `gorm:"type:text" json:"commit_message"`
	CommitAuthor  string     `gorm:"size:64" json:"commit_author"`
	CommitTime    *time.Time `gorm:"type:datetime" json:"commit_time"`
	// 外部CI集成
	ExternalPipelineID int64  `gorm:"column:external_pipeline_id" json:"external_pipeline_id"`
	ExternalWebURL     string `gorm:"column:external_web_url;size:255" json:"external_web_url"`
	// 审计时间戳与软删除
	CreatedAt time.Time      `gorm:"type:datetime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName 返回表名
func (Job) TableName() string { return "jobs" }
