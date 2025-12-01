package queue

import (
	"log"
	"sync"

	"gorm.io/gorm"
)

// queue 管理结构
var (
	jobQueue   chan uint64
	cancelLock sync.RWMutex
	cancelMap  map[uint64]bool
	globalDB   *gorm.DB
)

// InitWorkers 初始化队列与取消映射，并启动调度器
// 需在服务启动时调用
func InitWorkers(db *gorm.DB) {
	// 保存全局 DB 句柄供执行器使用
	globalDB = db
	// 有缓冲通道：允许短时间突发的任务提交，避免阻塞
	jobQueue = make(chan uint64, 128)
	// 取消映射：记录被请求取消的任务 ID
	cancelMap = make(map[uint64]bool)
	// 启动后台调度器：持续消费队列中的任务 ID
	go dispatcher()
	log.Printf("workers initialized")
}

// Enqueue 推入任务 ID 到队列
func Enqueue(id uint64) { jobQueue <- id }

// MarkCancel 标记取消任务
func MarkCancel(id uint64) {
	// 写锁保护：避免并发写入竞态
	cancelLock.Lock()
	cancelMap[id] = true
	cancelLock.Unlock()
}

// IsCanceled 判断任务是否取消
func IsCanceled(id uint64) bool {
	// 读锁保护：并发读取取消状态
	cancelLock.RLock()
	v := cancelMap[id]
	cancelLock.RUnlock()
	return v
}
