package gitlab

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
	"webci-refactored/internal/config"
	svc "webci-refactored/internal/service/gitlab"

	"github.com/xanzy/go-gitlab"
)

// Logic GitLab业务逻辑
type Logic struct {
	service *svc.Service
	config  config.Config
	mu      sync.Mutex
	hints   []taskHint
}

// NewLogic 创建GitLab业务逻辑实例
func NewLogic(cfg config.Config) (*Logic, error) {
	log.Printf("Creating GitLab logic with config: baseURL=%s, project=%s", cfg.GitLabBaseURL, cfg.GitLabProject)

	// 创建GitLab服务
	service, err := svc.NewService(cfg)
	if err != nil {
		log.Printf("Failed to create gitlab service: %v", err)
		return nil, err
	}

	return &Logic{
		service: service,
		config:  cfg,
	}, nil
}

func (l *Logic) UpdateConfig(cfg config.Config) error {
	if err := l.service.Reconfigure(cfg); err != nil {
		return err
	}
	l.config = cfg
	return nil
}

// ListPipelines 获取项目流水线列表
func (l *Logic) ListPipelines() ([]*PipelineInfo, error) {
	log.Printf("Logic: Listing pipelines")

	// 获取流水线列表
	pipelines, err := l.service.ListPipelines()
	if err != nil {
		log.Printf("Logic: Failed to list pipelines: %v", err)
		return nil, err
	}

	// 转换为内部结构
	var result []*PipelineInfo
	for _, p := range pipelines {
		result = append(result, &PipelineInfo{
			ID:     p.ID,
			Status: p.Status,
			Ref:    p.Ref,
			Sha:    p.SHA,
			WebURL: p.WebURL,
		})
	}

	log.Printf("Logic: Successfully converted %d pipelines", len(result))
	return result, nil
}

// GetPipelineDetails 获取流水线详细信息
func (l *Logic) GetPipelineDetails(pipelineID int) (*PipelineDetails, error) {
	log.Printf("Logic: Getting details for pipeline %d", pipelineID)

	// 获取流水线详情
	pipeline, err := l.service.GetPipeline(pipelineID)
	if err != nil {
		log.Printf("Logic: Failed to get pipeline %d: %v", pipelineID, err)
		return nil, err
	}

	// 获取作业列表
	jobs, err := l.service.ListJobs(pipelineID)
	if err != nil {
		log.Printf("Logic: Failed to list jobs for pipeline %d: %v", pipelineID, err)
		return nil, err
	}

	// 转换为内部结构
	var jobInfos []*JobInfo
	for _, j := range jobs {
		jobInfos = append(jobInfos, &JobInfo{
			ID:     j.ID,
			Name:   j.Name,
			Status: j.Status,
			Stage:  j.Stage,
		})
	}

	log.Printf("Logic: Successfully got details for pipeline %d with %d jobs", pipelineID, len(jobInfos))
	return &PipelineDetails{
		ID:     pipeline.ID,
		Status: pipeline.Status,
		Ref:    pipeline.Ref,
		Sha:    pipeline.SHA,
		Jobs:   jobInfos,
	}, nil
}

// ListBranches 获取项目分支列表
func (l *Logic) ListBranches() ([]*BranchInfo, error) {
	log.Printf("Logic: Listing branches")

	// 获取分支列表
	branches, err := l.service.GetBranches()
	if err != nil {
		log.Printf("Logic: Failed to list branches: %v", err)
		return nil, err
	}

	// 转换为内部结构
	var result []*BranchInfo
	for _, b := range branches {
		result = append(result, &BranchInfo{
			Name:      b.Name,
			CommitSHA: b.Commit.ID,
			Protected: b.Protected,
		})
	}

	log.Printf("Logic: Successfully converted %d branches", len(result))
	return result, nil
}

func (l *Logic) CreateBranch(name, ref string) (*BranchInfo, error) {
	b, err := l.service.CreateBranch(name, ref)
	if err != nil {
		return nil, err
	}
	return &BranchInfo{Name: b.Name, CommitSHA: b.Commit.ID, Protected: b.Protected}, nil
}

type CreateMRInput struct {
	SourceBranch string
	TargetBranch string
	Title        string
	Description  string
	Squash       bool
	RemoveSource bool
	MWPS         bool
}

func (l *Logic) CreateMergeRequest(in CreateMRInput) (*PipelineInfo, *JobInfo, *gitlab.MergeRequest, error) {
	mrs, err := l.service.ListOpenMergeRequestsBySourceTarget(in.SourceBranch, in.TargetBranch)
	if err == nil && len(mrs) > 0 {
		return nil, nil, mrs[0], nil
	}
	_ = l.service.EnsureBranch(in.TargetBranch, "main")
	mr, err := l.service.CreateMergeRequest(in.SourceBranch, in.TargetBranch, in.Title, in.Description, in.Squash, in.RemoveSource, in.MWPS)
	if err != nil {
		return nil, nil, nil, err
	}
	l.RecordMergeBranchHintWithURL(in.TargetBranch, mr.WebURL, in.Title)
	return nil, nil, mr, nil
}

type PromoteInput struct {
	SourcePrefix string
	Name         string
	Target       string
	Title        string
	Description  string
	Squash       bool
	RemoveSource bool
	MWPS         bool
}

func (l *Logic) Promote(in PromoteInput) (*gitlab.MergeRequest, error) {
	var source string
	var target string
	switch in.SourcePrefix {
	case "feature":
		source = "feature/" + in.Name
		if in.Target == "" || in.Target == "auto" {
			target = "test/" + in.Name
		} else {
			target = in.Target
		}
	case "test":
		source = "test/" + in.Name
		if in.Target == "" || in.Target == "auto" {
			target = "release/" + in.Name
		} else {
			target = in.Target
		}
	case "release":
		source = "release/" + in.Name
		if in.Target == "" || in.Target == "auto" {
			target = "main"
		} else {
			target = in.Target
		}
	default:
		return nil, fmt.Errorf("invalid source prefix")
	}
	mrs, err := l.service.ListOpenMergeRequestsBySourceTarget(source, target)
	if err == nil && len(mrs) > 0 {
		return mrs[0], nil
	}
	_ = l.service.EnsureBranch(target, "main")
	title := in.Title
	if title == "" {
		title = source + " -> " + target
	}
	mr, err := l.service.CreateMergeRequest(source, target, title, in.Description, in.Squash, in.RemoveSource, in.MWPS)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func (l *Logic) AcceptMergeRequest(iid int, squash, removeSource, mwps bool, message string) (*gitlab.MergeRequest, error) {
	if m, err := l.service.GetMergeRequest(iid); err == nil && m != nil {
		if m.State != "opened" {
			return nil, fmt.Errorf("merge request not opened: state=%s", m.State)
		}
		if m.WorkInProgress {
			return nil, fmt.Errorf("merge request is draft/WIP")
		}
		if m.HasConflicts {
			return nil, fmt.Errorf("merge request has conflicts")
		}
		if !m.BlockingDiscussionsResolved {
			return nil, fmt.Errorf("blocking discussions unresolved")
		}
		// 等待GitLab完成可合并性检查（checking/unchecked状态）
		status := m.MergeStatus
		deadline := time.Now().Add(20 * time.Second)
		for status == "checking" || status == "unchecked" || status == "cannot_be_merged_recheck" {
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("merge status=%s, cannot accept", status)
			}
			time.Sleep(1500 * time.Millisecond)
			mm, err := l.service.GetMergeRequest(iid)
			if err != nil || mm == nil {
				continue
			}
			status = mm.MergeStatus
			if mm.HasConflicts {
				return nil, fmt.Errorf("merge request has conflicts")
			}
		}
		if status != "can_be_merged" {
			return nil, fmt.Errorf("merge status=%s, cannot accept", status)
		}
	}
	mr, err := l.service.AcceptMergeRequest(iid, squash, removeSource, mwps, message)
	if err != nil {
		return nil, err
	}
	l.RecordMergeFromMR(mr)
	return mr, nil
}

// ListJobsForCIPage 获取用于CI模拟器页面展示的流水线任务列表
func (l *Logic) ListJobsForCIPage() ([]*GitLabJobInfo, error) {
	log.Printf("Logic: Listing jobs for CI page")

	// 获取流水线列表
	pipelines, err := l.service.ListPipelines()
	if err != nil {
		log.Printf("Logic: Failed to list pipelines: %v", err)
		return nil, err
	}

	// 转换为CI模拟器页面所需格式
	results := make([]*GitLabJobInfo, len(pipelines))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i, p := range pipelines {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, pi *gitlab.PipelineInfo) {
			defer func() { <-sem; wg.Done() }()
			d, err := l.service.GetPipeline(pi.ID)
			if err != nil {
				return
			}
			var duration string
			if d.Duration > 0 {
				duration = fmt.Sprintf("%ds", d.Duration)
			} else if d.CreatedAt != nil && d.UpdatedAt != nil {
				diff := d.UpdatedAt.Sub(*d.CreatedAt)
				duration = fmt.Sprintf("%ds", int(diff.Seconds()))
			}
			var createdAt string
			var createdTs int64
			if d.CreatedAt != nil {
				loc, _ := time.LoadLocation("Asia/Shanghai")
				createdAt = d.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
				createdTs = d.CreatedAt.In(loc).UnixMilli()
			}
			var triggerUser string
			if d.User != nil {
				if d.User.Name != "" {
					triggerUser = d.User.Name
				} else {
					triggerUser = d.User.Username
				}
			}
			var commitMsg string
			var commitAuthor string
			if pi.SHA != "" {
				if c, err := l.service.GetCommit(pi.SHA); err == nil && c != nil {
					commitMsg = c.Title
					if commitMsg == "" {
						commitMsg = c.Message
					}
					commitAuthor = c.AuthorName
				}
			}
			results[idx] = &GitLabJobInfo{ID: int64(pi.ID), Status: pi.Status, BranchName: pi.Ref, EnvironmentName: "", TriggerUser: triggerUser, CommitID: pi.SHA, CommitMessage: commitMsg, CommitAuthor: commitAuthor, CreatedAt: createdAt, WebURL: pi.WebURL, Duration: duration, TaskType: "修改文件"}
			_ = createdTs
		}(i, p)
	}
	wg.Wait()
	var jobs []*GitLabJobInfo
	seen := make(map[int64]struct{})
	for _, r := range results {
		if r == nil {
			continue
		}
		if _, ok := seen[r.ID]; ok {
			continue
		}
		seen[r.ID] = struct{}{}
		jobs = append(jobs, r)
	}

	l.applyTaskTypeClassification(jobs)
	log.Printf("Logic: Successfully converted %d jobs for CI page", len(jobs))
	return jobs, nil
}

func (l *Logic) ListJobsForCIPageWithPagination(page, perPage int) ([]*GitLabJobInfo, error) {
	pipelines, _, err := l.service.ListPipelinesWithPagination(page, perPage)
	if err != nil {
		return nil, err
	}
	results := make([]*GitLabJobInfo, len(pipelines))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i, p := range pipelines {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, pi *gitlab.PipelineInfo) {
			defer func() { <-sem; wg.Done() }()
			d, err := l.service.GetPipeline(pi.ID)
			if err != nil {
				return
			}
			var duration string
			if d.Duration > 0 {
				duration = fmt.Sprintf("%ds", d.Duration)
			} else if d.CreatedAt != nil && d.UpdatedAt != nil {
				diff := d.UpdatedAt.Sub(*d.CreatedAt)
				duration = fmt.Sprintf("%ds", int(diff.Seconds()))
			}
			var createdAt string
			var createdTs int64
			if d.CreatedAt != nil {
				loc, _ := time.LoadLocation("Asia/Shanghai")
				createdAt = d.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
				createdTs = d.CreatedAt.In(loc).UnixMilli()
			}
			var triggerUser string
			if d.User != nil {
				if d.User.Name != "" {
					triggerUser = d.User.Name
				} else {
					triggerUser = d.User.Username
				}
			}
			var commitMsg string
			var commitAuthor string
			if pi.SHA != "" {
				if c, err := l.service.GetCommit(pi.SHA); err == nil && c != nil {
					commitMsg = c.Title
					if commitMsg == "" {
						commitMsg = c.Message
					}
					commitAuthor = c.AuthorName
				}
			}
			results[idx] = &GitLabJobInfo{ID: int64(pi.ID), Status: pi.Status, BranchName: pi.Ref, EnvironmentName: "", TriggerUser: triggerUser, CommitID: pi.SHA, CommitMessage: commitMsg, CommitAuthor: commitAuthor, CreatedAt: createdAt, WebURL: pi.WebURL, Duration: duration, TaskType: "修改文件"}
			_ = createdTs
		}(i, p)
	}
	wg.Wait()
	var jobs []*GitLabJobInfo
	seen := make(map[int64]struct{})
	for _, r := range results {
		if r == nil {
			continue
		}
		if _, ok := seen[r.ID]; ok {
			continue
		}
		seen[r.ID] = struct{}{}
		jobs = append(jobs, r)
	}
	l.applyTaskTypeClassification(jobs)
	return jobs, nil
}

func (l *Logic) applyTaskTypeClassification(jobs []*GitLabJobInfo) {
	l.mu.Lock()
	hints := append([]taskHint(nil), l.hints...)
	l.mu.Unlock()
	byBranch := make(map[string][]*GitLabJobInfo)
	for _, j := range jobs {
		b := j.BranchName
		byBranch[b] = append(byBranch[b], j)
		j.TaskType = "修改文件"
	}
	// 排序按创建时间升序
	loc, _ := time.LoadLocation("Asia/Shanghai")
	parse := func(s string) time.Time { t, _ := time.ParseInLocation("2006-01-02 15:04:05", s, loc); return t }
	for b, arr := range byBranch {
		if len(arr) == 0 {
			continue
		}
		sort.Slice(arr, func(i, k int) bool { return parse(arr[i].CreatedAt).Before(parse(arr[k].CreatedAt)) })
		// 处理创建分支：将提示后的第一条记录钉为创建分支
		for _, h := range hints {
			if h.Kind == "create_branch" && h.Branch == b {
				var best *GitLabJobInfo
				for _, j := range arr {
					ct := parse(j.CreatedAt)
					if h.Ts.IsZero() || !ct.IsZero() {
						if ct.After(h.Ts.Add(-2 * time.Minute)) {
							best = j
							break
						}
						if best == nil {
							best = j
						}
					}
				}
				if best != nil {
					best.TaskType = "创建分支"
					best.CommitMessage = "创建分支"
				}
			}
		}
		// 处理合并：在提示时间窗口内标记为分支合并（不覆盖创建分支）
		for _, h := range hints {
			if h.Kind == "merge" && h.Branch == b {
				if h.SHA != "" {
					var matched bool
					for _, j := range arr {
						if j.TaskType == "创建分支" {
							continue
						}
						if j.CommitID == h.SHA {
							j.TaskType = "分支合并"
							matched = true
						}
					}
					// 若有SHA但未匹配到任何记录，则追加一个合成记录以展示此次合并
					if !matched {
						created := h.Ts.In(loc).Format("2006-01-02 15:04:05")
						synth := &GitLabJobInfo{ID: -time.Now().UnixNano(), Status: "pending", BranchName: h.Branch, EnvironmentName: "", TriggerUser: "", CommitID: h.SHA, CommitMessage: h.Title, CommitAuthor: "", CreatedAt: created, WebURL: h.URL, Duration: "", TaskType: "分支合并"}
						jobs = append(jobs, synth)
					}
				} else {
					// 没有合并提交SHA（例如仅创建了MR），生成一条合成记录用于列表展示，不影响其它记录分类
					created := h.Ts.In(loc).Format("2006-01-02 15:04:05")
					synth := &GitLabJobInfo{ID: -time.Now().UnixNano(), Status: "pending", BranchName: h.Branch, EnvironmentName: "", TriggerUser: "", CommitID: "", CommitMessage: h.Title, CommitAuthor: "", CreatedAt: created, WebURL: h.URL, Duration: "", TaskType: "分支合并"}
					jobs = append(jobs, synth)
				}
			}
		}
	}
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
	NextPage    int `json:"next_page"`
	PrevPage    int `json:"prev_page"`
}

type GitLabJobsPage struct {
	Items      []*GitLabJobInfo `json:"items"`
	Pagination Pagination       `json:"pagination"`
}

func (l *Logic) ListJobsPage(page, perPage int) (*GitLabJobsPage, error) {
	_, resp, err := l.service.ListPipelinesWithPagination(page, perPage)
	if err != nil {
		return nil, err
	}
	items, err := l.ListJobsForCIPageWithPagination(page, perPage)
	if err != nil {
		return nil, err
	}
	cur := resp.CurrentPage
	tot := resp.TotalPages
	var next int
	var prev int
	if tot > 0 {
		if cur < tot {
			next = cur + 1
		}
		if cur > 1 {
			prev = cur - 1
		}
	}
	pg := Pagination{CurrentPage: cur, PerPage: perPage, TotalPages: tot, TotalItems: 0, NextPage: next, PrevPage: prev}
	return &GitLabJobsPage{Items: items, Pagination: pg}, nil
}

// PipelineInfo 流水线信息
type PipelineInfo struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Ref    string `json:"ref"`
	Sha    string `json:"sha"`
	WebURL string `json:"web_url"`
}

// JobInfo 作业信息
type JobInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Stage  string `json:"stage"`
}

// PipelineDetails 流水线详细信息
type PipelineDetails struct {
	ID     int        `json:"id"`
	Status string     `json:"status"`
	Ref    string     `json:"ref"`
	Sha    string     `json:"sha"`
	Jobs   []*JobInfo `json:"jobs"`
}

// BranchInfo 分支信息
type BranchInfo struct {
	Name      string `json:"name"`
	CommitSHA string `json:"commit_sha"`
	Protected bool   `json:"protected"`
}

// GitLabJobInfo GitLab流水线任务信息（用于CI模拟器页面展示）
type GitLabJobInfo struct {
	ID              int64  `json:"id"`
	Status          string `json:"status"`
	BranchName      string `json:"branch_name"`
	EnvironmentName string `json:"environment_name"`
	TriggerUser     string `json:"trigger_user"`
	CommitID        string `json:"commit_id"`
	CommitMessage   string `json:"commit_message"`
	CommitAuthor    string `json:"commit_author"`
	CreatedAt       string `json:"created_at"`
	WebURL          string `json:"web_url"`
	Duration        string `json:"duration"`
	TaskType        string `json:"task_type"`
}
type taskHint struct {
	Kind   string
	Branch string
	Ts     time.Time
	URL    string
	Title  string
	SHA    string
}

func (l *Logic) RecordCreateBranchHint(branch string) {
	l.mu.Lock()
	l.hints = append(l.hints, taskHint{Kind: "create_branch", Branch: branch, Ts: time.Now()})
	l.mu.Unlock()
}

func (l *Logic) RecordMergeBranchHint(branch string) {
	l.mu.Lock()
	l.hints = append(l.hints, taskHint{Kind: "merge", Branch: branch, Ts: time.Now()})
	l.mu.Unlock()
}

func (l *Logic) RecordMergeBranchHintWithURL(branch, url, title string) {
	l.mu.Lock()
	l.hints = append(l.hints, taskHint{Kind: "merge", Branch: branch, Ts: time.Now(), URL: url, Title: title})
	l.mu.Unlock()
}

func (l *Logic) RecordMergeFromMR(mr *gitlab.MergeRequest) {
	if mr == nil {
		return
	}
	ts := time.Now()
	if mr.MergedAt != nil {
		ts = *mr.MergedAt
	}
	l.mu.Lock()
	l.hints = append(l.hints, taskHint{Kind: "merge", Branch: mr.TargetBranch, Ts: ts, URL: mr.WebURL, Title: mr.Title, SHA: mr.MergeCommitSHA})
	l.mu.Unlock()
}
