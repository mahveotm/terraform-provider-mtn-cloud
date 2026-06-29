package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// captureServer records the POST body's wrapped object and echoes a minimal
// object with an id back under the same key.
func captureServer(t *testing.T, wantPath, key string, capture *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		*capture = payload[key].(map[string]any)
		echo := map[string]any{"id": 7}
		_ = json.NewEncoder(w).Encode(map[string]any{key: echo})
	}))
}

func govClient(t *testing.T, url string) *Client {
	t.Helper()
	c, err := New(Config{URL: url, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRoleBodyMergesPermissionSet(t *testing.T) {
	t.Parallel()
	var role map[string]any
	server := captureServer(t, "/api/roles", "role", &role)
	defer server.Close()

	if _, err := govClient(t, server.URL).CreateRole(context.Background(), RoleInput{
		Name:          "ops",
		Description:   "Operators",
		RoleType:      "user",
		PermissionSet: `{"globalSiteAccess":"all","featurePermissions":[{"code":"admin-users","access":"full"}]}`,
	}); err != nil {
		t.Fatal(err)
	}

	if role["authority"] != "ops" || role["roleType"] != "user" {
		t.Fatalf("unexpected role base: %#v", role)
	}
	if role["globalSiteAccess"] != "all" {
		t.Fatalf("permission_set globalSiteAccess not merged: %#v", role)
	}
	fp, ok := role["featurePermissions"].([]any)
	if !ok || len(fp) != 1 {
		t.Fatalf("permission_set featurePermissions not merged: %#v", role["featurePermissions"])
	}
}

func TestUserBodyRoles(t *testing.T) {
	t.Parallel()
	var user map[string]any
	server := captureServer(t, "/api/users", "user", &user)
	defer server.Close()

	if _, err := govClient(t, server.URL).CreateUser(context.Background(), UserInput{
		Username: "jdoe",
		Email:    "jdoe@example.com",
		Password: "secret",
		RoleIDs:  []int64{3, 5},
	}); err != nil {
		t.Fatal(err)
	}

	if user["username"] != "jdoe" || user["email"] != "jdoe@example.com" || user["password"] != "secret" {
		t.Fatalf("unexpected user base: %#v", user)
	}
	roles, ok := user["roles"].([]any)
	if !ok || len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %#v", user["roles"])
	}
	first := roles[0].(map[string]any)
	if first["id"].(float64) != 3 {
		t.Fatalf("unexpected first role id: %#v", first)
	}
}

func TestUserGroupBody(t *testing.T) {
	t.Parallel()
	var ug map[string]any
	server := captureServer(t, "/api/user-groups", "userGroup", &ug)
	defer server.Close()

	sudo := true
	if _, err := govClient(t, server.URL).CreateUserGroup(context.Background(), UserGroupInput{
		Name:       "platform",
		SudoAccess: &sudo,
		UserIDs:    []int64{9},
	}); err != nil {
		t.Fatal(err)
	}

	if ug["name"] != "platform" || ug["sudoUser"] != true {
		t.Fatalf("unexpected user group base: %#v", ug)
	}
	users, ok := ug["users"].([]any)
	if !ok || len(users) != 1 || users[0].(map[string]any)["id"].(float64) != 9 {
		t.Fatalf("unexpected users: %#v", ug["users"])
	}
}
