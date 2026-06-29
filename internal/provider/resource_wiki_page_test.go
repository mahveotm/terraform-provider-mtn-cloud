package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestWikiContentSemanticEqualsTrailingNewline(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	current := newWikiContentValue("# Deployment Runbook\n")
	stored := newWikiContentValue("# Deployment Runbook")

	equal, diags := current.StringSemanticEquals(ctx, stored)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !equal {
		t.Fatal("expected values that differ by one trailing newline to be semantically equal")
	}
}

func TestWikiContentSemanticEqualsDetectsContentChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	current := newWikiContentValue("# Deployment Runbook\n")
	stored := newWikiContentValue("# Different Runbook")

	equal, diags := current.StringSemanticEquals(ctx, stored)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if equal {
		t.Fatal("expected different content to be semantically different")
	}
}

func TestWikiContentSemanticEqualsKnownValuesOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	current := wikiContentValue{StringValue: basetypes.NewStringUnknown()}
	stored := newWikiContentValue("# Deployment Runbook")

	equal, diags := current.StringSemanticEquals(ctx, stored)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if equal {
		t.Fatal("expected unknown content to be semantically different")
	}

	equal, diags = newWikiContentValue("# Deployment Runbook").StringSemanticEquals(ctx, types.StringNull())
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if equal {
		t.Fatal("expected null content to be semantically different")
	}
}
