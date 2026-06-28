package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateNetworkPayloadShape(t *testing.T) {
	t.Parallel()

	typeID := int64(45)
	poolID := int64(214)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/networks" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		network := payload["network"].(map[string]any)
		if network["name"] != "app-net" {
			t.Fatalf("unexpected name: %#v", network)
		}
		if network["site"].(map[string]any)["id"].(float64) != 621 {
			t.Fatalf("unexpected site: %#v", network["site"])
		}
		if network["zone"].(map[string]any)["id"].(float64) != 4 {
			t.Fatalf("unexpected zone: %#v", network["zone"])
		}
		if network["type"].(map[string]any)["id"].(float64) != 45 {
			t.Fatalf("unexpected type: %#v", network["type"])
		}
		if network["zonePool"].(map[string]any)["id"].(float64) != 214 {
			t.Fatalf("unexpected zonePool: %#v", network["zonePool"])
		}
		if network["cidr"] != "10.42.10.0/24" || network["gateway"] != "10.42.10.1" {
			t.Fatalf("unexpected cidr/gateway: %#v", network)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"network": map[string]any{"id": 99, "name": "app-net"}})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	created, err := c.CreateNetwork(context.Background(), NetworkInput{
		Name:           "app-net",
		GroupID:        621,
		CloudID:        4,
		TypeID:         &typeID,
		ResourcePoolID: &poolID,
		CIDR:           "10.42.10.0/24",
		Gateway:        "10.42.10.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 99 {
		t.Fatalf("unexpected created network: %#v", created)
	}
}

func TestGetNetworkTypeByNameFiltersOpenStack(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/network-types" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"networkTypes": []map[string]any{
				{"id": 1, "name": "Internal", "code": "internal", "category": "standard"},
				{"id": 45, "name": "OpenStack Private", "code": "openstackPrivate", "category": "openstack"},
			},
		})
	}))
	defer server.Close()

	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	networkType, err := c.GetNetworkTypeByName(context.Background(), "OpenStack Private", true)
	if err != nil {
		t.Fatal(err)
	}
	if networkType.ID != 45 {
		t.Fatalf("unexpected network type: %#v", networkType)
	}

	if _, err := c.GetNetworkTypeByName(context.Background(), "internal", true); !IsNotFound(err) {
		t.Fatalf("expected non-openstack type to be filtered out, got %v", err)
	}
}
