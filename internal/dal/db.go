package dal

import (
	"webci-refactored/internal/config"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB 初始化 GORM 数据库连接
// 使用 MySQL 驱动，根据配置中的 DSN 建立连接
func InitDB(cfg config.Config) (*gorm.DB, error) {
	// 当未提供 MySQL DSN 时，走内存 SQLite：无需安装数据库，适合快速演示与练习
	// 说明：使用 shared cache 让多个连接共享同一内存数据库
	if cfg.MySQLDSN == "" {
		return gorm.Open(sqlite.Open("file:webci_demo?mode=memory&cache=shared"), &gorm.Config{})
	}
	// 提供了 DSN 时，使用 MySQL 作为持久化存储
	return gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
}
