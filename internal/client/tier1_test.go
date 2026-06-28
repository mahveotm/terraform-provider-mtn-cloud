package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMapCredentialPayloadByType(t *testing.T) {
	t.Parallel()

	kp := int64(7)
	cases := []struct {
		name   string
		input  CredentialInput
		expect map[string]any // expected keys inside "credential"
	}{
		{"access-key-secret", CredentialInput{Type: "access-key-secret", Name: "aws", AccessKey: "AK", SecretKey: "SK"},
			map[string]any{"type": "access-key-secret", "username": "AK", "password": "SK"}},
		{"api-key", CredentialInput{Type: "api-key", Name: "k", APIKey: "abc"},
			map[string]any{"password": "abc"}},
		{"username-password", CredentialInput{Type: "username-password", Name: "u", Username: "bob", Password: "pw"},
			map[string]any{"username": "bob", "password": "pw"}},
		{"username-keypair", CredentialInput{Type: "username-keypair", Name: "u", Username: "bob", KeyPairID: &kp},
			map[string]any{"username": "bob"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := mapCredentialPayload(tc.input)
			cred := body["credential"].(map[string]any)
			for k, v := range tc.expect {
				if cred[k] != v {
					t.Fatalf("%s: expected %s=%v, got %v (full=%#v)", tc.name, k, v, cred[k], cred)
				}
			}
			if tc.input.KeyPairID != nil {
				kpMap, ok := cred["keyPair"].(map[string]any)
				if !ok || kpMap["id"] != *tc.input.KeyPairID {
					t.Fatalf("%s: expected keyPair.id=%d, got %#v", tc.name, *tc.input.KeyPairID, cred["keyPair"])
				}
			}
		})
	}
}

func TestCreateIPPoolSendsRanges(t *testing.T) {
	t.Parallel()

	c, closeFn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/networks/pools" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		pool := payload["networkPool"].(map[string]any)
		if pool["type"] != "morpheus" || pool["name"] != "pool-a" {
			t.Fatalf("unexpected pool: %#v", pool)
		}
		ranges := pool["ipRanges"].([]any)
		first := ranges[0].(map[string]any)
		if first["startAddress"] != "10.0.0.10" || first["endAddress"] != "10.0.0.20" {
			t.Fatalf("unexpected range: %#v", first)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"networkPool": map[string]any{"id": 9, "name": "pool-a"}})
	})
	defer closeFn()

	pool, err := c.CreateIPPool(context.Background(), IPPoolInput{
		Name:     "pool-a",
		IPRanges: []IPRange{{StartAddress: "10.0.0.10", EndAddress: "10.0.0.20"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pool.ID != 9 {
		t.Fatalf("unexpected pool: %#v", pool)
	}
}

func TestBudgetBodySetsPeriodAndCosts(t *testing.T) {
	t.Parallel()

	enabled := true
	body := budgetBody(BudgetInput{Name: "q-budget", Interval: "quarter", Year: "2026", Currency: "NGN", Enabled: &enabled, Costs: []float64{100, 200, 300, 400}})
	budget := body["budget"].(map[string]any)
	if budget["period"] != "year" || budget["interval"] != "quarter" {
		t.Fatalf("unexpected period/interval: %#v", budget)
	}
	costs := budget["costs"].([]float64)
	if len(costs) != 4 || costs[3] != 400 {
		t.Fatalf("unexpected costs: %#v", costs)
	}
	if budget["currency"] != "NGN" || budget["year"] != "2026" {
		t.Fatalf("unexpected currency/year: %#v", budget)
	}
}

func TestScaleThresholdBodyOmitsUnsetPointers(t *testing.T) {
	t.Parallel()

	up := true
	maxCPU := 80.0
	body := scaleThresholdBody(ScaleThresholdInput{Name: "cpu", AutoUp: &up, MaxCPU: &maxCPU})
	st := body["scaleThreshold"].(map[string]any)
	if st["name"] != "cpu" || st["autoUp"] != true || st["maxCpu"] != 80.0 {
		t.Fatalf("unexpected scaleThreshold: %#v", st)
	}
	if _, ok := st["minCpu"]; ok {
		t.Fatalf("expected unset minCpu to be omitted: %#v", st)
	}
	if _, ok := st["autoDown"]; ok {
		t.Fatalf("expected unset autoDown to be omitted: %#v", st)
	}
}

func TestNetworkDomainBodyOmitsEmpty(t *testing.T) {
	t.Parallel()

	pub := true
	body := networkDomainBody(NetworkDomainInput{Name: "corp.local", Visibility: "private", PublicZone: &pub})
	domain := body["networkDomain"].(map[string]any)
	if domain["name"] != "corp.local" || domain["visibility"] != "private" || domain["publicZone"] != true {
		t.Fatalf("unexpected domain: %#v", domain)
	}
	if _, ok := domain["fqdn"]; ok {
		t.Fatalf("expected empty fqdn to be omitted: %#v", domain)
	}
}
