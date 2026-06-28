package client

import (
	"context"
	"fmt"
)

// Environment is a deployment environment (e.g. dev/stage/prod) in MTN Cloud.
type Environment struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Visibility  string `json:"visibility"`
	Active      *bool  `json:"active"`
}

// EnvironmentInput is the create/update payload.
type EnvironmentInput struct {
	Name        string
	Description string
	Code        string
	Visibility  string
	Active      *bool
}

func environmentBody(input EnvironmentInput) map[string]any {
	env := map[string]any{"name": input.Name}
	if input.Description != "" {
		env["description"] = input.Description
	}
	if input.Code != "" {
		env["code"] = input.Code
	}
	if input.Visibility != "" {
		env["visibility"] = input.Visibility
	}
	if input.Active != nil {
		env["active"] = *input.Active
	}
	return map[string]any{"environment": env}
}

func (c *Client) CreateEnvironment(ctx context.Context, input EnvironmentInput) (*Environment, error) {
	return createObj[Environment](c, ctx, "/environments", "environment", environmentBody(input))
}

func (c *Client) GetEnvironment(ctx context.Context, id int64) (*Environment, error) {
	return getByID[Environment](c, ctx, fmt.Sprintf("/environments/%d", id), "environment")
}

func (c *Client) GetEnvironmentByName(ctx context.Context, name string) (*Environment, error) {
	return firstByName[Environment](c, ctx, "/environments", "environments", name)
}

func (c *Client) UpdateEnvironment(ctx context.Context, id int64, input EnvironmentInput) (*Environment, error) {
	return updateObj[Environment](c, ctx, fmt.Sprintf("/environments/%d", id), "environment", environmentBody(input))
}

func (c *Client) DeleteEnvironment(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/environments/%d", id), nil)
}

func (c *Client) ListEnvironments(ctx context.Context) ([]Environment, error) {
	return listObjects[Environment](c, ctx, "/environments", "environments")
}
