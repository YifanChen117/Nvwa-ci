package config

import (
	"os"
)

// Config 应用配置
// 包含 HTTP 监听地址、MySQL DSN、Git 仓库路径
type Config struct {
	HTTPAddr      string
	MySQLDSN      string
	RepoPath      string
	GitLabBaseURL string
	GitLabToken   string
	GitLabProject string
}

// Load 读取环境变量生成配置
// 默认 HTTPAddr=:8080，MySQLDSN 从 MYSQL_DSN，RepoPath 从 REPO_PATH
func Load() Config {
	// HTTP 监听地址：为空则使用 :8080（本机所有网卡）
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	// MySQL 连接串：留空表示使用内存 SQLite（演示友好）
	dsn := os.Getenv("MYSQL_DSN")
	// Git 仓库路径：用于分支 refresh 功能（可选）
	repo := os.Getenv("REPO_PATH")
	// GitLab 配置：从环境变量读取
	glURL := os.Getenv("GITLAB_BASE_URL")
	glToken := os.Getenv("GITLAB_TOKEN")
	glProj := os.Getenv("GITLAB_PROJECT_ID")
	return Config{HTTPAddr: addr, MySQLDSN: dsn, RepoPath: repo, GitLabBaseURL: glURL, GitLabToken: glToken, GitLabProject: glProj}
}
