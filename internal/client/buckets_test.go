package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateStorageBucketEmbedsConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/storage-buckets" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		bucket := payload["storageBucket"].(map[string]any)
		if bucket["providerType"] != "s3" || bucket["bucketName"] != "my-archive" {
			t.Fatalf("unexpected bucket: %#v", bucket)
		}
		config := bucket["config"].(map[string]any)
		if config["accessKey"] != "AK" || config["secretKey"] != "SK" || config["endpoint"] != "https://s3.example" {
			t.Fatalf("unexpected config: %#v", config)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"storageBucket": map[string]any{"id": 7, "name": "store"}})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	created, err := c.CreateStorageBucket(context.Background(), StorageBucketInput{
		Name:       "store",
		BucketName: "my-archive",
		AccessKey:  "AK",
		SecretKey:  "SK",
		Endpoint:   "https://s3.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 7 {
		t.Fatalf("unexpected created bucket: %#v", created)
	}
}

func TestCreateArchiveBucketPayloadShape(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/archives/buckets" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		bucket := payload["archiveBucket"].(map[string]any)
		if bucket["name"] != "vault" {
			t.Fatalf("unexpected name: %#v", bucket)
		}
		if bucket["storageProvider"].(map[string]any)["id"].(float64) != 7 {
			t.Fatalf("unexpected storageProvider: %#v", bucket["storageProvider"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"archiveBucket": map[string]any{"id": 12, "name": "vault"}})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	created, err := c.CreateArchiveBucket(context.Background(), ArchiveBucketInput{
		Name:              "vault",
		StorageProviderID: 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 12 {
		t.Fatalf("unexpected created archive bucket: %#v", created)
	}
}
