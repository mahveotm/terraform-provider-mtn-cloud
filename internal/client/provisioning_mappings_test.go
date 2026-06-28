package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Regression tests for the live MTN Cloud API field/endpoint mappings used by
// provisioning resolution. Each guards a bug where the client read a field the
// API does not actually populate. See memory: terraform-provider-provisioning-field-mappings.

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	c, err := New(Config{URL: server.URL, Token: "test-token", Timeout: time.Second})
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return c, server.Close
}

// Group clouds come from the embedded "zones" array, not a flat "cloudIds".
func TestGroupCloudIDsDerivedFromZones(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/groups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groups": []map[string]any{
				{
					"id":   621,
					"name": "MTNNG_CLOUD_AZ_1",
					"zones": []map[string]any{
						{"id": 4, "name": "MTNNG_CLOUD_AZ_1"},
					},
				},
			},
		})
	})
	defer closeFn()

	group, err := c.GetGroupByName(context.Background(), "MTNNG_CLOUD_AZ_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(group.CloudIDs) != 1 || group.CloudIDs[0] != 4 {
		t.Fatalf("expected CloudIDs [4] from zones, got %v", group.CloudIDs)
	}
}

// An explicit cloudIds field is still honored when present.
func TestGroupCloudIDsHonorsExplicitField(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groups": []map[string]any{
				{"id": 1, "name": "g", "cloudIds": []int{9}, "zones": []map[string]any{{"id": 4}}},
			},
		})
	})
	defer closeFn()

	group, err := c.GetGroupByName(context.Background(), "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(group.CloudIDs) != 1 || group.CloudIDs[0] != 9 {
		t.Fatalf("expected explicit CloudIDs [9], got %v", group.CloudIDs)
	}
}

// Layout id comes from instanceTypeLayouts[0].id, not the null defaultLayoutId.
func TestInstanceTypeLayoutFromInstanceTypeLayouts(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/instance-types" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"instanceTypes": []map[string]any{
				{
					"id":              104,
					"code":            "MTN-CS10",
					"defaultLayoutId": nil,
					"instanceTypeLayouts": []map[string]any{
						{"id": 327, "name": "MTN-CentOS Stream 10"},
					},
				},
			},
		})
	})
	defer closeFn()

	it, err := c.GetInstanceTypeByCode(context.Background(), "MTN-CS10")
	if err != nil {
		t.Fatal(err)
	}
	if it.DefaultLayoutID == nil || *it.DefaultLayoutID != 327 {
		t.Fatalf("expected layout 327 from instanceTypeLayouts, got %v", it.DefaultLayoutID)
	}
}

// Resource pool code comes from "value" (e.g. pool-214); group rows are skipped.
func TestResourcePoolCodeFromValue(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/options/zonePools" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("cloudId") != "4" || r.URL.Query().Get("groupId") != "621" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"name": "Resource Pools", "isGroup": true},
				{"id": 214, "name": "fourthischarm-Marv-Osuolale", "value": "pool-214"},
			},
		})
	})
	defer closeFn()

	group := &Group{ID: 621, Name: "MTNNG_CLOUD_AZ_1", CloudIDs: []int64{4}}
	pool, err := c.GetResourcePool(context.Background(), "fourthischarm-Marv-Osuolale", group)
	if err != nil {
		t.Fatal(err)
	}
	if pool.ID != 214 || pool.Code != "pool-214" {
		t.Fatalf("expected pool 214/pool-214, got id=%d code=%q", pool.ID, pool.Code)
	}
}

// Service plans are read from the /options/servicePlans option-source.
func TestServicePlansFromOptionSource(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/options/servicePlans" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("zoneId") != "4" || r.URL.Query().Get("layoutId") != "327" || r.URL.Query().Get("siteId") != "621" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": 9082, "name": "G1S1"},
				{"id": 9112, "name": "G1S2"},
			},
		})
	})
	defer closeFn()

	plan, err := c.GetServicePlan(context.Background(), "G1S1", 4, 327, 621)
	if err != nil {
		t.Fatal(err)
	}
	if plan.ID != 9082 {
		t.Fatalf("expected plan id 9082, got %d", plan.ID)
	}
}
