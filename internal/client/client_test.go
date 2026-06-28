package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetGroupByNameUsesBearerTokenAndQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/groups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("name") != "MTNNG_CLOUD_AZ_1" || r.URL.Query().Get("max") != "1" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groups": []map[string]any{
				{"id": 621, "name": "MTNNG_CLOUD_AZ_1", "cloudIds": []int{4}, "location": "Lagos", "active": true},
			},
		})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	group, err := c.GetGroupByName(context.Background(), "MTNNG_CLOUD_AZ_1")
	if err != nil {
		t.Fatal(err)
	}
	if group.ID != 621 || group.CloudIDs[0] != 4 {
		t.Fatalf("unexpected group: %#v", group)
	}
}

func TestAuthenticateWithUsernamePassword(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			if r.URL.Query().Get("grant_type") != "password" || r.URL.Query().Get("client_id") != "morph-cli" {
				t.Fatalf("unexpected oauth query: %s", r.URL.RawQuery)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("username") != "user@example.com" || r.Form.Get("password") != "secret" {
				t.Fatalf("unexpected form: %#v", r.Form)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "oauth-token"})
		case "/api/groups":
			if r.Header.Get("Authorization") != "Bearer oauth-token" {
				t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"groups": []map[string]any{{"id": 1, "name": "group", "cloudIds": []int{2}}}})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Username: "user@example.com", Password: "secret", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetGroupByName(context.Background(), "group"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateInstancePayloadMatchesProvisioningShape(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/instances" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		instance := payload["instance"].(map[string]any)
		config := payload["config"].(map[string]any)
		if instance["name"] != "web-01" || instance["cloud"] != "MTNNG_CLOUD_AZ_1" {
			t.Fatalf("unexpected instance payload: %#v", instance)
		}
		if instance["instanceType"].(map[string]any)["code"] != "MTN-CS10" {
			t.Fatalf("missing instanceType code: %#v", instance)
		}
		if config["resourcePoolId"] != "pool-214" || config["securityGroup"] != "web" {
			t.Fatalf("unexpected config payload: %#v", config)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"instance": map[string]any{"id": 123, "name": "web-01", "status": "provisioning"}})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	created, err := c.ProvisionInstance(context.Background(), CreateInstanceInput{
		Name:           "web-01",
		Cloud:          "MTNNG_CLOUD_AZ_1",
		Type:           "MTN-CS10",
		GroupID:        621,
		LayoutID:       327,
		PlanID:         6923,
		ResourcePoolID: "pool-214",
		SecurityGroup:  "web",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 123 {
		t.Fatalf("unexpected created instance: %#v", created)
	}
}

func TestErrorMappingIncludesStatusCode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "not found"})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetGroupByName(context.Background(), "missing")
	if !IsNotFound(err) {
		t.Fatalf("expected not found APIError, got %T %[1]v", err)
	}
}
