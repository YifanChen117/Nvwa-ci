package provider

import (
    "context"
    "webci-refactored/sdk/types"
)

type VCSProvider interface {
    CreateBranch(ctx context.Context, name string, baseRef string) (*types.Branch, error)
    ListBranches(ctx context.Context) ([]*types.Branch, error)
    CreateMergeRequest(ctx context.Context, in types.CreateMRInput) (*types.MergeRequest, error)
    AcceptMergeRequest(ctx context.Context, iid int, opts types.AcceptMROptions) (*types.MergeRequest, error)
    GetMergeRequest(ctx context.Context, iid int) (*types.MergeRequest, error)
    ListPipelines(ctx context.Context, opts types.PipelineListOptions) ([]*types.Pipeline, error)
    ListJobs(ctx context.Context, pipelineID int) ([]*types.Job, error)
    GetCommit(ctx context.Context, sha string) (*types.Commit, error)
}
