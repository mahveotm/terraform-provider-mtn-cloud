package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateKeyPairPayloadShape(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/key-pairs" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		kp := payload["keyPair"].(map[string]any)
		if kp["name"] != "deploy" || kp["publicKey"] != "ssh-rsa AAAA" {
			t.Fatalf("unexpected keyPair: %#v", kp)
		}
		if kp["privateKey"] != "PRIVATE" || kp["passphrase"] != "pass" {
			t.Fatalf("unexpected private fields: %#v", kp)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keyPair": map[string]any{"id": 5, "name": "deploy", "publicKey": "ssh-rsa AAAA"}})
	})
	defer closeFn()

	created, err := c.CreateKeyPair(context.Background(), KeyPairInput{
		Name:       "deploy",
		PublicKey:  "ssh-rsa AAAA",
		PrivateKey: "PRIVATE",
		Passphrase: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 5 || created.PublicKey != "ssh-rsa AAAA" {
		t.Fatalf("unexpected created key pair: %#v", created)
	}
}

func TestGetKeyPairByNameUsesListWrapper(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/key-pairs" || r.URL.Query().Get("name") != "deploy" {
			t.Fatalf("unexpected request: %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keyPairs": []map[string]any{{"id": 5, "name": "deploy"}}})
	})
	defer closeFn()

	kp, err := c.GetKeyPairByName(context.Background(), "deploy")
	if err != nil {
		t.Fatal(err)
	}
	if kp.ID != 5 {
		t.Fatalf("unexpected key pair: %#v", kp)
	}
}

func TestCreateCypherSendsTTLAndType(t *testing.T) {
	t.Parallel()

	ttl := int64(3600)
	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cypher/secret/myapp/db-password" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("ttl") != "3600" || r.URL.Query().Get("type") != "string" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["value"] != "s3cr3t" {
			t.Fatalf("unexpected value: %#v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"cypher":         map[string]any{"id": 7, "itemKey": "secret/myapp/db-password"},
			"lease_duration": 3600,
		})
	})
	defer closeFn()

	result, err := c.CreateCypher(context.Background(), "myapp/db-password", "s3cr3t", &ttl)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cypher.ID != 7 || result.LeaseDuration == nil || *result.LeaseDuration != 3600 {
		t.Fatalf("unexpected cypher result: %#v", result)
	}
}

func TestGetCypherReturnsDecryptedValue(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cypher/secret/myapp/db-password" || r.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"cypher":         map[string]any{"id": 7, "itemKey": "secret/myapp/db-password"},
			"data":           "s3cr3t",
			"lease_duration": 1800,
		})
	})
	defer closeFn()

	result, err := c.GetCypher(context.Background(), "myapp/db-password")
	if err != nil {
		t.Fatal(err)
	}
	if result.Value != "s3cr3t" || result.LeaseDuration == nil || *result.LeaseDuration != 1800 {
		t.Fatalf("unexpected cypher result: %#v", result)
	}
}

func TestCreateWikiPageTrimsTrailingNewline(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/wiki/pages" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		page := payload["page"].(map[string]any)
		if page["content"] != "line one\nline two" {
			t.Fatalf("expected trailing newline trimmed, got %#v", page["content"])
		}
		if page["name"] != "Runbook" || page["category"] != "ops" {
			t.Fatalf("unexpected page: %#v", page)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"page": map[string]any{"id": 3, "name": "Runbook", "category": "ops", "content": "line one\nline two"}})
	})
	defer closeFn()

	page, err := c.CreateWikiPage(context.Background(), WikiPageInput{
		Name:     "Runbook",
		Category: "ops",
		Content:  "line one\nline two\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.ID != 3 {
		t.Fatalf("unexpected wiki page: %#v", page)
	}
}

func TestEnvironmentBodyOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	active := true
	body := environmentBody(EnvironmentInput{Name: "prod", Visibility: "private", Active: &active})
	env := body["environment"].(map[string]any)
	if env["name"] != "prod" || env["visibility"] != "private" || env["active"] != true {
		t.Fatalf("unexpected environment body: %#v", env)
	}
	if _, ok := env["description"]; ok {
		t.Fatalf("expected empty description to be omitted: %#v", env)
	}
	if _, ok := env["code"]; ok {
		t.Fatalf("expected empty code to be omitted: %#v", env)
	}
}
