package client

import (
    pgl "webci-refactored/sdk/provider/gitlab"
)

func NewGitLabClient(token, baseURL, projectID string) (*Client, error) {
    p, err := pgl.New(token, baseURL, projectID)
    if err != nil {
        return nil, err
    }
    return New(p), nil
}
