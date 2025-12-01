package client

import (
    "context"
    "webci-refactored/sdk/provider"
    "webci-refactored/sdk/types"
)

type Client struct {
    provider provider.VCSProvider
}

func New(provider provider.VCSProvider) *Client {
    return &Client{provider: provider}
}

func (c *Client) CreateBranch(ctx context.Context, name, baseRef string) (*types.Branch, error) {
    return c.provider.CreateBranch(ctx, name, baseRef)
}

func (c *Client) ListBranches(ctx context.Context) ([]*types.Branch, error) {
    return c.provider.ListBranches(ctx)
}

func (c *Client) CreateMergeRequest(ctx context.Context, in types.CreateMRInput) (*types.MergeRequest, error) {
    return c.provider.CreateMergeRequest(ctx, in)
}

func (c *Client) AcceptMergeRequest(ctx context.Context, iid int, opts types.AcceptMROptions) (*types.MergeRequest, error) {
    return c.provider.AcceptMergeRequest(ctx, iid, opts)
}

func (c *Client) GetMergeRequest(ctx context.Context, iid int) (*types.MergeRequest, error) {
    return c.provider.GetMergeRequest(ctx, iid)
}

func (c *Client) ListPipelines(ctx context.Context, opts types.PipelineListOptions) ([]*types.Pipeline, error) {
    return c.provider.ListPipelines(ctx, opts)
}

func (c *Client) ListJobs(ctx context.Context, pipelineID int) ([]*types.Job, error) {
    return c.provider.ListJobs(ctx, pipelineID)
}

func (c *Client) GetCommit(ctx context.Context, sha string) (*types.Commit, error) {
    return c.provider.GetCommit(ctx, sha)
}
