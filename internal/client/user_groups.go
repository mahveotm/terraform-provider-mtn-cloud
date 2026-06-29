package client

import (
	"context"
	"fmt"
	"sort"
)

// UserGroup is an MTN Cloud user group.
type UserGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SudoUser    *bool  `json:"sudoUser"`
	ServerGroup string `json:"serverGroup"`
	Users       []struct {
		ID int64 `json:"id"`
	} `json:"users"`
}

// UserIDList returns the group's member user IDs, sorted for stable comparison.
func (g *UserGroup) UserIDList() []int64 {
	ids := make([]int64, 0, len(g.Users))
	for _, u := range g.Users {
		ids = append(ids, u.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// UserGroupInput is the create/update payload.
type UserGroupInput struct {
	Name        string
	Description string
	SudoAccess  *bool
	ServerGroup string
	UserIDs     []int64
}

func userGroupBody(in UserGroupInput) map[string]any {
	ug := map[string]any{"name": in.Name}
	setIf(ug, "description", in.Description)
	setIf(ug, "serverGroup", in.ServerGroup)
	if in.SudoAccess != nil {
		ug["sudoUser"] = *in.SudoAccess
	}
	users := make([]map[string]any, 0, len(in.UserIDs))
	for _, id := range in.UserIDs {
		users = append(users, map[string]any{"id": id})
	}
	ug["users"] = users
	return map[string]any{"userGroup": ug}
}

func (c *Client) CreateUserGroup(ctx context.Context, input UserGroupInput) (*UserGroup, error) {
	return createObj[UserGroup](c, ctx, "/user-groups", "userGroup", userGroupBody(input))
}

func (c *Client) GetUserGroup(ctx context.Context, id int64) (*UserGroup, error) {
	return getByID[UserGroup](c, ctx, fmt.Sprintf("/user-groups/%d", id), "userGroup")
}

// GetUserGroupByName lists user groups and matches on `name`.
func (c *Client) GetUserGroupByName(ctx context.Context, name string) (*UserGroup, error) {
	groups, err := c.ListUserGroups(ctx)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("user group %q not found", name)}
}

func (c *Client) UpdateUserGroup(ctx context.Context, id int64, input UserGroupInput) (*UserGroup, error) {
	return updateObj[UserGroup](c, ctx, fmt.Sprintf("/user-groups/%d", id), "userGroup", userGroupBody(input))
}

func (c *Client) DeleteUserGroup(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/user-groups/%d", id), nil)
}

func (c *Client) ListUserGroups(ctx context.Context) ([]UserGroup, error) {
	return listObjects[UserGroup](c, ctx, "/user-groups", "userGroups")
}
