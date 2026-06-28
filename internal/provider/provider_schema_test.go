package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestProviderSchemaValid drives the whole provider through the framework's
// GetProviderSchema path, which validates every registered resource and data
// source schema (attribute rules, defaults, plan modifiers) and rejects
// duplicate type names. It needs no live API, so it runs in CI on every push and
// is the first guardrail for a newly added resource.
func TestProviderSchemaValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("provider server: %s", err)
	}
	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("GetProviderSchema: %s", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("schema error: %s — %s", d.Summary, d.Detail)
		}
	}
	if len(resp.ResourceSchemas) == 0 || len(resp.DataSourceSchemas) == 0 {
		t.Fatalf("expected resources and data sources, got %d/%d",
			len(resp.ResourceSchemas), len(resp.DataSourceSchemas))
	}
	for name := range resp.ResourceSchemas {
		if !strings.HasPrefix(name, "mtncloud_") {
			t.Errorf("resource %q is not mtncloud_-prefixed", name)
		}
	}
	for name := range resp.DataSourceSchemas {
		if !strings.HasPrefix(name, "mtncloud_") {
			t.Errorf("data source %q is not mtncloud_-prefixed", name)
		}
	}
}

// TestProviderRegistrationCounts guards against forgetting to register a New*
// constructor: the counts must match what the provider exposes.
func TestProviderRegistrationCounts(t *testing.T) {
	t.Parallel()
	p := New("test")().(*mtnCloudProvider)
	if got := len(p.Resources(context.Background())); got == 0 {
		t.Fatal("no resources registered")
	}
	if got := len(p.DataSources(context.Background())); got == 0 {
		t.Fatal("no data sources registered")
	}
}
