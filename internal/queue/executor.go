package queue

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"webci-refactored/internal/config"
	"webci-refactored/internal/dal/model"
	"webci-refactored/internal/dal/repository"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// dispatcher 任务调度器
// 从队列中取出 jobID 并执行 Mock 构建
func dispatcher() {
	repo := repository.NewJobRepository(globalDB)
	branches := repository.NewBranchRepository(globalDB)
	envs := repository.NewEnvironmentRepository(globalDB)
	cfg := config.Load()
	for id := range jobQueue {
		go func(jobID uint64) {
			// 并发执行：每个任务一个 goroutine，模拟构建流水线
			log.Printf("dispatch job=%d", jobID)
			now := time.Now()
			// 状态切换为 running，并记录开始时间
			if err := repo.UpdateStatus(jobID, "running"); err != nil {
				log.Printf("status running err: %v", err)
				return
			}
			if err := repo.UpdateTimes(jobID, &now, nil); err != nil {
				log.Printf("start time err: %v", err)
			}
			_ = repo.AppendLog(jobID, fmt.Sprintf("[START] job %d at %s\n", jobID, now.Format(time.RFC3339)))
			// 模拟执行耗时：2~5 秒随机
			dur := time.Duration(2+rand.Intn(4)) * time.Second
			_ = repo.AppendLog(jobID, fmt.Sprintf("[RUNNING] executing for %s...\n", dur))
			time.Sleep(dur)
			end := time.Now()
			// 取消优先：被标记取消的任务直接失败并记录日志
			if IsCanceled(jobID) {
				_ = repo.UpdateTimes(jobID, nil, &end)
				_ = repo.UpdateStatus(jobID, "failed")
				_ = repo.AppendLog(jobID, "[FAILED] job cancelled\n")
				return
			}
			// 随机成功/失败：70% 成功率用于演示
			ok := rand.Intn(100) >= 30
			_ = repo.UpdateTimes(jobID, nil, &end)
			if ok {
				_ = repo.UpdateStatus(jobID, "success")
				_ = repo.AppendLog(jobID, "[SUCCESS] job finished\n")
				j, _ := repo.Get(jobID)
				if cfg.RepoPath != "" {
					if r, err := git.PlainOpen(cfg.RepoPath); err == nil {
						b, _ := branches.Get(j.BranchID)
						if b != nil {
							ref := plumbing.ReferenceName("refs/heads/" + b.Name)
							if rf, err2 := r.Reference(ref, true); err2 == nil {
								if co, err3 := r.CommitObject(rf.Hash()); err3 == nil {
									ct := co.Author.When
									globalDB.Model(&model.Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
										"commit_id":      rf.Hash().String(),
										"commit_message": co.Message,
										"commit_author":  co.Author.Name,
										"commit_time":    ct,
									})
								}
							}
						}
					}
				}
				if j.EnvID > 0 {
					e, _ := envs.Get(j.EnvID)
					if e != nil {
						t := time.Now()
						globalDB.Model(&model.Environment{}).Where("id = ?", j.EnvID).Updates(map[string]interface{}{
							"current_deploy_commit": j.CommitID,
							"current_deploy_at":     t,
						})
					}
				}
			} else {
				_ = repo.UpdateStatus(jobID, "failed")
				_ = repo.AppendLog(jobID, "[FAILED] build failed\n")
			}
		}(id)
	}
}
