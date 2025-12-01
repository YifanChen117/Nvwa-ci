package gitlab

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"webci-refactored/internal/config"

	"github.com/xanzy/go-gitlab"
)

// Service GitLab服务
type Service struct {
	client        *gitlab.Client
	config        config.Config
	mu            sync.Mutex
	cacheTTL      time.Duration
	pipelineCache map[int]struct {
		v   *gitlab.Pipeline
		exp time.Time
	}
	commitCache map[string]struct {
		v   *gitlab.Commit
		exp time.Time
	}
}

// NewService 创建GitLab服务实例
func NewService(cfg config.Config) (*Service, error) {
	// 验证必要配置
	if cfg.GitLabToken == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN is required")
	}
	if cfg.GitLabBaseURL == "" {
		return nil, fmt.Errorf("GITLAB_BASE_URL is required")
	}
	if cfg.GitLabProject == "" {
		return nil, fmt.Errorf("GITLAB_PROJECT_ID is required")
	}

	log.Printf("Creating GitLab client with baseURL: %s, project: %s", cfg.GitLabBaseURL, cfg.GitLabProject)

	// 创建自定义HTTP客户端，跳过SSL证书验证（仅用于测试环境）
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: transport}

	// 创建GitLab客户端
	client, err := gitlab.NewClient(cfg.GitLabToken,
		gitlab.WithBaseURL(cfg.GitLabBaseURL),
		gitlab.WithHTTPClient(httpClient))
	if err != nil {
		log.Printf("Failed to create gitlab client: %v", err)
		return nil, fmt.Errorf("failed to create gitlab client: %v", err)
	}

	s := &Service{client: client, config: cfg}
	s.cacheTTL = 30 * time.Second
	s.pipelineCache = make(map[int]struct {
		v   *gitlab.Pipeline
		exp time.Time
	})
	s.commitCache = make(map[string]struct {
		v   *gitlab.Commit
		exp time.Time
	})
	return s, nil
}

func (s *Service) Reconfigure(cfg config.Config) error {
	if cfg.GitLabToken == "" {
		return fmt.Errorf("GITLAB_TOKEN is required")
	}
	if cfg.GitLabBaseURL == "" {
		return fmt.Errorf("GITLAB_BASE_URL is required")
	}
	if cfg.GitLabProject == "" {
		return fmt.Errorf("GITLAB_PROJECT_ID is required")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{Transport: transport}
	client, err := gitlab.NewClient(cfg.GitLabToken, gitlab.WithBaseURL(cfg.GitLabBaseURL), gitlab.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("failed to create gitlab client: %v", err)
	}
	s.client = client
	s.config = cfg
	return nil
}

// ListPipelines 获取项目流水线列表
func (s *Service) ListPipelines() ([]*gitlab.PipelineInfo, error) {
	log.Printf("Listing pipelines for project: %s", s.config.GitLabProject)

	// 获取流水线列表
	pipelines, resp, err := s.client.Pipelines.ListProjectPipelines(s.config.GitLabProject, &gitlab.ListProjectPipelinesOptions{
		Sort: gitlab.String("desc"),
	})

	if err != nil {
		log.Printf("Failed to list pipelines: %v, response: %+v", err, resp)
		return nil, fmt.Errorf("failed to list pipelines: %v", err)
	}

	log.Printf("Successfully listed %d pipelines", len(pipelines))
	return pipelines, nil
}

func (s *Service) ListPipelinesWithPagination(page, perPage int) ([]*gitlab.PipelineInfo, *gitlab.Response, error) {
	log.Printf("Listing pipelines (page=%d, perPage=%d) for project: %s", page, perPage, s.config.GitLabProject)
	pipelines, resp, err := s.client.Pipelines.ListProjectPipelines(s.config.GitLabProject, &gitlab.ListProjectPipelinesOptions{
		Sort:        gitlab.String("desc"),
		ListOptions: gitlab.ListOptions{Page: page, PerPage: perPage},
	})
	if err != nil {
		log.Printf("Failed to list pipelines: %v, response: %+v", err, resp)
		return nil, resp, fmt.Errorf("failed to list pipelines: %v", err)
	}
	return pipelines, resp, nil
}

// GetPipeline 获取单个流水线详情
func (s *Service) GetPipeline(pipelineID int) (*gitlab.Pipeline, error) {
	s.mu.Lock()
	if e, ok := s.pipelineCache[pipelineID]; ok && time.Now().Before(e.exp) {
		v := e.v
		s.mu.Unlock()
		return v, nil
	}
	s.mu.Unlock()
	log.Printf("Getting pipeline %d for project: %s", pipelineID, s.config.GitLabProject)
	pipeline, resp, err := s.client.Pipelines.GetPipeline(s.config.GitLabProject, pipelineID)
	if err != nil {
		log.Printf("Failed to get pipeline %d: %v, response: %+v", pipelineID, err, resp)
		return nil, fmt.Errorf("failed to get pipeline: %v", err)
	}
	s.mu.Lock()
	s.pipelineCache[pipelineID] = struct {
		v   *gitlab.Pipeline
		exp time.Time
	}{v: pipeline, exp: time.Now().Add(s.cacheTTL)}
	s.mu.Unlock()
	return pipeline, nil
}

// ListJobs 获取流水线作业列表
func (s *Service) ListJobs(pipelineID int) ([]*gitlab.Job, error) {
	log.Printf("Listing jobs for pipeline %d in project: %s", pipelineID, s.config.GitLabProject)

	// 获取作业列表
	jobs, resp, err := s.client.Jobs.ListPipelineJobs(s.config.GitLabProject, pipelineID, &gitlab.ListJobsOptions{})
	if err != nil {
		log.Printf("Failed to list jobs for pipeline %d: %v, response: %+v", pipelineID, err, resp)
		return nil, fmt.Errorf("failed to list jobs: %v", err)
	}

	log.Printf("Successfully listed %d jobs for pipeline %d", len(jobs), pipelineID)
	return jobs, nil
}

// GetBranches 获取项目分支列表
func (s *Service) GetBranches() ([]*gitlab.Branch, error) {
	log.Printf("Listing branches for project: %s", s.config.GitLabProject)

	// 获取分支列表
	branches, resp, err := s.client.Branches.ListBranches(s.config.GitLabProject, &gitlab.ListBranchesOptions{})
	if err != nil {
		log.Printf("Failed to list branches: %v, response: %+v", err, resp)
		return nil, fmt.Errorf("failed to list branches: %v", err)
	}

	log.Printf("Successfully listed %d branches", len(branches))
	return branches, nil
}

// GetCommits 获取项目提交列表
func (s *Service) GetCommits() ([]*gitlab.Commit, error) {
	log.Printf("Listing commits for project: %s", s.config.GitLabProject)

	// 获取提交列表
	commits, resp, err := s.client.Commits.ListCommits(s.config.GitLabProject, &gitlab.ListCommitsOptions{})
	if err != nil {
		log.Printf("Failed to list commits: %v, response: %+v", err, resp)
		return nil, fmt.Errorf("failed to list commits: %v", err)
	}

	log.Printf("Successfully listed %d commits", len(commits))
	return commits, nil
}

func (s *Service) GetCommit(sha string) (*gitlab.Commit, error) {
	s.mu.Lock()
	if e, ok := s.commitCache[sha]; ok && time.Now().Before(e.exp) {
		v := e.v
		s.mu.Unlock()
		return v, nil
	}
	s.mu.Unlock()
	commit, resp, err := s.client.Commits.GetCommit(s.config.GitLabProject, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %v", err)
	}
	_ = resp
	s.mu.Lock()
	s.commitCache[sha] = struct {
		v   *gitlab.Commit
		exp time.Time
	}{v: commit, exp: time.Now().Add(s.cacheTTL)}
	s.mu.Unlock()
	return commit, nil
}

func (s *Service) CreateBranch(name, ref string) (*gitlab.Branch, error) {
	opt := &gitlab.CreateBranchOptions{
		Branch: gitlab.String(name),
		Ref:    gitlab.String(ref),
	}
	b, resp, err := s.client.Branches.CreateBranch(s.config.GitLabProject, opt)
	if err != nil {
		log.Printf("Failed to create branch: %v, response: %+v", err, resp)
		return nil, fmt.Errorf("failed to create branch: %v", err)
	}
	return b, nil
}

func (s *Service) CreateMergeRequest(source, target, title, description string, squash, removeSource, mwps bool) (*gitlab.MergeRequest, error) {
	opt := &gitlab.CreateMergeRequestOptions{
		SourceBranch:       gitlab.String(source),
		TargetBranch:       gitlab.String(target),
		Title:              gitlab.String(title),
		Description:        nil,
		RemoveSourceBranch: gitlab.Bool(removeSource),
		Squash:             gitlab.Bool(squash),
	}
	if description != "" {
		opt.Description = gitlab.String(description)
	}
	mr, resp, err := s.client.MergeRequests.CreateMergeRequest(s.config.GitLabProject, opt)
	if err != nil {
		log.Printf("Failed to create merge request: %v, response: %+v", err, resp)
		return nil, fmt.Errorf("failed to create merge request: %v", err)
	}
	return mr, nil
}

func (s *Service) ListOpenMergeRequestsBySourceTarget(source, target string) ([]*gitlab.MergeRequest, error) {
	opt := &gitlab.ListProjectMergeRequestsOptions{State: gitlab.String("opened"), SourceBranch: gitlab.String(source), TargetBranch: gitlab.String(target)}
	mrs, resp, err := s.client.MergeRequests.ListProjectMergeRequests(s.config.GitLabProject, opt)
	if err != nil {
		_ = resp
		return nil, fmt.Errorf("failed to list merge requests: %v", err)
	}
	return mrs, nil
}

func (s *Service) BranchExists(name string) (bool, error) {
	branches, err := s.GetBranches()
	if err != nil {
		return false, err
	}
	for _, b := range branches {
		if b.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) EnsureBranch(name, ref string) error {
	exists, err := s.BranchExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = s.CreateBranch(name, ref)
	return err
}

func (s *Service) AcceptMergeRequest(iid int, squash, removeSource, mwps bool, message string) (*gitlab.MergeRequest, error) {
	opt := &gitlab.AcceptMergeRequestOptions{
		MergeCommitMessage:        nil,
		ShouldRemoveSourceBranch:  gitlab.Bool(removeSource),
		MergeWhenPipelineSucceeds: gitlab.Bool(mwps),
		Squash:                    gitlab.Bool(squash),
	}
	if message != "" {
		opt.MergeCommitMessage = gitlab.String(message)
	}
	mr, resp, err := s.client.MergeRequests.AcceptMergeRequest(s.config.GitLabProject, iid, opt)
	if err != nil {
		log.Printf("Failed to accept merge request: %v, response: %+v", err, resp)
		return nil, fmt.Errorf("failed to accept merge request: %v", err)
	}
	return mr, nil
}

func (s *Service) GetMergeRequest(iid int) (*gitlab.MergeRequest, error) {
	mr, resp, err := s.client.MergeRequests.GetMergeRequest(s.config.GitLabProject, iid, nil)
	if err != nil {
		_ = resp
		return nil, fmt.Errorf("failed to get merge request: %v", err)
	}
	return mr, nil
}
