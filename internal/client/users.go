package client

import (
	"context"
	"fmt"
	"sort"
)

// User is an MTN Cloud user account. The password is write-only (never returned).
// Cross-tenant placement is not supported here: managing other accounts needs
// admin-accounts access, which the customer-admin token lacks.
type User struct {
	ID                   int64  `json:"id"`
	Username             string `json:"username"`
	Email                string `json:"email"`
	FirstName            string `json:"firstName"`
	LastName             string `json:"lastName"`
	ReceiveNotifications *bool  `json:"receiveNotifications"`
	Roles                []struct {
		ID int64 `json:"id"`
	} `json:"roles"`
}

// RoleIDList returns the user's role IDs, sorted for stable comparison.
func (u *User) RoleIDList() []int64 {
	ids := make([]int64, 0, len(u.Roles))
	for _, r := range u.Roles {
		ids = append(ids, r.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// UserInput is the create/update payload.
type UserInput struct {
	Username             string
	Email                string
	Password             string
	FirstName            string
	LastName             string
	RoleIDs              []int64
	PasswordExpired      *bool
	ReceiveNotifications *bool
	LinuxUsername        string
	LinuxPassword        string
	LinuxKeyPairID       *int64
	WindowsUsername      string
	WindowsPassword      string
}

func userBody(in UserInput) map[string]any {
	user := map[string]any{"username": in.Username, "email": in.Email}
	setIf(user, "firstName", in.FirstName)
	setIf(user, "lastName", in.LastName)
	setIf(user, "password", in.Password)
	setIf(user, "linuxUsername", in.LinuxUsername)
	setIf(user, "linuxPassword", in.LinuxPassword)
	setIf(user, "windowsUsername", in.WindowsUsername)
	setIf(user, "windowsPassword", in.WindowsPassword)
	if in.PasswordExpired != nil {
		user["passwordExpired"] = *in.PasswordExpired
	}
	if in.ReceiveNotifications != nil {
		user["receiveNotifications"] = *in.ReceiveNotifications
	}
	if in.LinuxKeyPairID != nil {
		user["linuxKeyPairId"] = *in.LinuxKeyPairID
	}
	roles := make([]map[string]any, 0, len(in.RoleIDs))
	for _, id := range in.RoleIDs {
		roles = append(roles, map[string]any{"id": id})
	}
	user["roles"] = roles
	return map[string]any{"user": user}
}

func (c *Client) CreateUser(ctx context.Context, input UserInput) (*User, error) {
	return createObj[User](c, ctx, "/users", "user", userBody(input))
}

func (c *Client) GetUser(ctx context.Context, id int64) (*User, error) {
	return getByID[User](c, ctx, fmt.Sprintf("/users/%d", id), "user")
}

// GetUserByName lists users and matches on `username` (the API's `?name=` filter
// does not apply to users).
func (c *Client) GetUserByName(ctx context.Context, username string) (*User, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].Username == username {
			return &users[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("user %q not found", username)}
}

func (c *Client) UpdateUser(ctx context.Context, id int64, input UserInput) (*User, error) {
	return updateObj[User](c, ctx, fmt.Sprintf("/users/%d", id), "user", userBody(input))
}

func (c *Client) DeleteUser(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/users/%d", id), nil)
}

func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	return listObjects[User](c, ctx, "/users", "users")
}
