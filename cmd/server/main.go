package main

import (
	"log"

	"webci-refactored/internal/config"
	"webci-refactored/internal/dal"
	"webci-refactored/internal/queue"
	"webci-refactored/internal/router"
)

// main 启动应用入口
// 加载配置、初始化数据库与迁移、启动任务队列与 HTTP 服务
func main() {
	// 1) 加载配置：从环境变量读取 HTTP_ADDR / MYSQL_DSN / REPO_PATH
	cfg := config.Load()
	log.Printf("starting webci on %s", cfg.HTTPAddr)

	// 2) 初始化数据库连接：当 MYSQL_DSN 为空时使用内存 SQLite，便于演示
	db, err := dal.InitDB(cfg)
	if err != nil {
		log.Fatalf("init db error: %v", err)
	}

	// 3) 自动迁移：确保表结构与模型一致
	if err := dal.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate error: %v", err)
	}

	// 4) 初始化任务队列与执行器：提供入队、取消标记与并发执行能力
	queue.InitWorkers(db)

	// 5) 创建并运行 HTTP 服务（CloudWeGo Hertz）
	h := router.NewServer(cfg, db)
	if err := h.Run(); err != nil {
		log.Fatalf("server run error: %v", err)
	}
}
