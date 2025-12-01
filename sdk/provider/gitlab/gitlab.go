package gitlab

import (
	"context"
	"crypto/tls"
	"net/http"
	"webci-refactored/sdk/types"

	gl "github.com/xanzy/go-gitlab"
)

type GitLabProvider struct {
	client    *gl.Client
	projectID string
}

func New(token, baseURL, projectID string) (*GitLabProvider, error) {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{Transport: transport}
	c, err := gl.NewClient(token, gl.WithBaseURL(baseURL), gl.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &GitLabProvider{client: c, projectID: projectID}, nil
}

func (p *GitLabProvider) CreateBranch(ctx context.Context, name string, baseRef string) (*types.Branch, error) {
	opt := &gl.CreateBranchOptions{Branch: gl.String(name), Ref: gl.String(baseRef)}
	b, _, err := p.client.Branches.CreateBranch(p.projectID, opt, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &types.Branch{Name: b.Name, CommitSHA: b.Commit.ID, Protected: b.Protected}, nil
}

func (p *GitLabProvider) ListBranches(ctx context.Context) ([]*types.Branch, error) {
	bs, _, err := p.client.Branches.ListBranches(p.projectID, &gl.ListBranchesOptions{}, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]*types.Branch, 0, len(bs))
	for _, b := range bs {
		out = append(out, &types.Branch{Name: b.Name, CommitSHA: b.Commit.ID, Protected: b.Protected})
	}
	return out, nil
}

func (p *GitLabProvider) CreateMergeRequest(ctx context.Context, in types.CreateMRInput) (*types.MergeRequest, error) {
	opt := &gl.CreateMergeRequestOptions{
		SourceBranch:       gl.String(in.SourceBranch),
		TargetBranch:       gl.String(in.TargetBranch),
		Title:              gl.String(in.Title),
		RemoveSourceBranch: gl.Bool(in.RemoveSource),
		Squash:             gl.Bool(in.Squash),
	}
	if in.Description != "" {
		opt.Description = gl.String(in.Description)
	}
	mr, _, err := p.client.MergeRequests.CreateMergeRequest(p.projectID, opt, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return toMR(mr), nil
}

func (p *GitLabProvider) AcceptMergeRequest(ctx context.Context, iid int, opts types.AcceptMROptions) (*types.MergeRequest, error) {
	opt := &gl.AcceptMergeRequestOptions{
		ShouldRemoveSourceBranch:  gl.Bool(opts.RemoveSourceBranch),
		MergeWhenPipelineSucceeds: gl.Bool(opts.MergeWhenPipelineSucceeds),
		Squash:                    gl.Bool(opts.Squash),
	}
	if opts.MergeCommitMessage != "" {
		opt.MergeCommitMessage = gl.String(opts.MergeCommitMessage)
	}
	mr, _, err := p.client.MergeRequests.AcceptMergeRequest(p.projectID, iid, opt, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return toMR(mr), nil
}

func (p *GitLabProvider) GetMergeRequest(ctx context.Context, iid int) (*types.MergeRequest, error) {
	mr, _, err := p.client.MergeRequests.GetMergeRequest(p.projectID, iid, &gl.GetMergeRequestsOptions{}, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return toMR(mr), nil
}

func (p *GitLabProvider) ListPipelines(ctx context.Context, opts types.PipelineListOptions) ([]*types.Pipeline, error) {
	pOpts := &gl.ListProjectPipelinesOptions{Sort: gl.String("desc")}
	if opts.Page > 0 || opts.PerPage > 0 {
		pOpts.ListOptions = gl.ListOptions{Page: opts.Page, PerPage: opts.PerPage}
	}
	ps, _, err := p.client.Pipelines.ListProjectPipelines(p.projectID, pOpts, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]*types.Pipeline, 0, len(ps))
	for _, pi := range ps {
		out = append(out, &types.Pipeline{ID: pi.ID, Status: pi.Status, Ref: pi.Ref, SHA: pi.SHA, WebURL: pi.WebURL})
	}
	return out, nil
}

func (p *GitLabProvider) ListJobs(ctx context.Context, pipelineID int) ([]*types.Job, error) {
	js, _, err := p.client.Jobs.ListPipelineJobs(p.projectID, pipelineID, &gl.ListJobsOptions{}, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]*types.Job, 0, len(js))
	for _, j := range js {
		out = append(out, &types.Job{ID: j.ID, Name: j.Name, Status: j.Status, Stage: j.Stage})
	}
	return out, nil
}

func (p *GitLabProvider) GetCommit(ctx context.Context, sha string) (*types.Commit, error) {
	cm, _, err := p.client.Commits.GetCommit(p.projectID, sha, nil, gl.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &types.Commit{ID: cm.ID, ShortID: cm.ShortID, Title: cm.Title, Message: cm.Message, AuthorName: cm.AuthorName, AuthorEmail: cm.AuthorEmail, CreatedAt: cm.CreatedAt.String()}, nil
}

func toMR(m *gl.MergeRequest) *types.MergeRequest {
	if m == nil {
		return nil
	}
	return &types.MergeRequest{
		IID:                         m.IID,
		State:                       m.State,
		Title:                       m.Title,
		SourceBranch:                m.SourceBranch,
		TargetBranch:                m.TargetBranch,
		MergeStatus:                 m.MergeStatus,
		HasConflicts:                m.HasConflicts,
		WorkInProgress:              m.WorkInProgress,
		BlockingDiscussionsResolved: m.BlockingDiscussionsResolved,
	}
}
