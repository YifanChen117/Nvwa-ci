package client

import (
	"context"
	"os"
	"testing"
	"webci-refactored/sdk/types"
)

func TestGitLabClientListBranches(t *testing.T) {
	base := os.Getenv("GITLAB_BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")
	proj := os.Getenv("GITLAB_PROJECT_ID")
	if base == "" || token == "" || proj == "" {
		t.Skip("missing gitlab env")
	}
	c, err := NewGitLabClient(token, base, proj)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	bs, err := c.ListBranches(context.Background())
	if err != nil {
		t.Fatalf("list branches: %v", err)
	}
	if bs == nil {
		t.Fatalf("nil branches")
	}
}

func TestGitLabClientListPipelines(t *testing.T) {
	base := os.Getenv("GITLAB_BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")
	proj := os.Getenv("GITLAB_PROJECT_ID")
	if base == "" || token == "" || proj == "" {
		t.Skip("missing gitlab env")
	}
	c, err := NewGitLabClient(token, base, proj)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ps, err := c.ListPipelines(context.Background(), types.PipelineListOptions{Page: 1, PerPage: 5})
	if err != nil {
		t.Fatalf("list pipelines: %v", err)
	}
	if ps == nil {
		t.Fatalf("nil pipelines")
	}
}
