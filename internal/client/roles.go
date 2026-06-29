package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Role is an MTN Cloud (Morpheus) role. The role's name is carried by the API
// `authority` field. The large permission structure the API returns alongside a
// role is not decoded here; permission_set is config-authoritative (see the
// provider resource).
type Role struct {
	ID                int64  `json:"id"`
	Authority         string `json:"authority"`
	Description       string `json:"description"`
	RoleType          string `json:"roleType"`
	Multitenant       *bool  `json:"multitenant"`
	MultitenantLocked *bool  `json:"multitenantLocked"`
}

// RoleTypes are the role types the provider accepts. "account" (tenant) roles
// need admin-accounts access, which the customer-admin token lacks, so creating
// them is expected to fail; "user" roles are the supported default.
var RoleTypes = []string{"user", "account"}

// RoleInput is the create/update payload. PermissionSet is a raw JSON document in
// the Morpheus role API shape (e.g. {"globalSiteAccess":"all","featurePermissions":
// [{"code":"admin-users","access":"full"}]}); its keys are merged into the role
// object exactly as the API expects.
type RoleInput struct {
	Name              string
	Description       string
	RoleType          string
	Multitenant       *bool
	MultitenantLocked *bool
	PermissionSet     string
}

func roleBody(in RoleInput) map[string]any {
	role := map[string]any{"authority": in.Name}
	if in.Description != "" {
		role["description"] = in.Description
	}
	if in.RoleType != "" {
		role["roleType"] = in.RoleType
	}
	if in.Multitenant != nil {
		role["multitenant"] = *in.Multitenant
	}
	if in.MultitenantLocked != nil {
		role["multitenantLocked"] = *in.MultitenantLocked
	}
	if in.PermissionSet != "" {
		var perms map[string]any
		if err := json.Unmarshal([]byte(in.PermissionSet), &perms); err == nil {
			for k, v := range perms {
				role[k] = v
			}
		}
	}
	return map[string]any{"role": role}
}

func (c *Client) CreateRole(ctx context.Context, input RoleInput) (*Role, error) {
	return createObj[Role](c, ctx, "/roles", "role", roleBody(input))
}

func (c *Client) GetRole(ctx context.Context, id int64) (*Role, error) {
	return getByID[Role](c, ctx, fmt.Sprintf("/roles/%d", id), "role")
}

// GetRoleByName lists roles and matches on `authority` (the API's `?name=` filter
// does not apply to roles).
func (c *Client) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	roles, err := c.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].Authority == name {
			return &roles[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("role %q not found", name)}
}

func (c *Client) UpdateRole(ctx context.Context, id int64, input RoleInput) (*Role, error) {
	return updateObj[Role](c, ctx, fmt.Sprintf("/roles/%d", id), "role", roleBody(input))
}

func (c *Client) DeleteRole(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/roles/%d", id), nil)
}

func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	return listObjects[Role](c, ctx, "/roles", "roles")
}
