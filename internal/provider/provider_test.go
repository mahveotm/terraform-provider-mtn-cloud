package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	p := New("test")()
	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "mtncloud" {
		t.Fatalf("unexpected provider type name: %s", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Fatalf("unexpected provider version: %s", resp.Version)
	}
}

func TestProviderSchemaIncludesSensitiveAuthFields(t *testing.T) {
	t.Parallel()

	p := New("test")()
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	token, ok := resp.Schema.Attributes["token"]
	if !ok || !token.IsSensitive() {
		t.Fatalf("token attribute missing or not sensitive")
	}
	password, ok := resp.Schema.Attributes["password"]
	if !ok || !password.IsSensitive() {
		t.Fatalf("password attribute missing or not sensitive")
	}
}
