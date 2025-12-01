package types

type Branch struct {
    Name      string
    CommitSHA string
    Protected bool
}

type MergeRequest struct {
    IID                        int
    State                      string
    Title                      string
    SourceBranch               string
    TargetBranch               string
    MergeStatus                string
    HasConflicts               bool
    WorkInProgress             bool
    BlockingDiscussionsResolved bool
}

type Pipeline struct {
    ID     int
    Status string
    Ref    string
    SHA    string
    WebURL string
}

type Job struct {
    ID     int
    Name   string
    Status string
    Stage  string
}

type Commit struct {
    ID          string
    ShortID     string
    Title       string
    Message     string
    AuthorName  string
    AuthorEmail string
    CreatedAt   string
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

type AcceptMROptions struct {
    Squash                 bool
    RemoveSourceBranch     bool
    MergeWhenPipelineSucceeds bool
    MergeCommitMessage     string
}

type PipelineListOptions struct {
    Page    int
    PerPage int
}
