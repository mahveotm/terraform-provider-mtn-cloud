package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runStringValidator(v validator.String, value types.String) bool {
	var resp validator.StringResponse
	v.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("test"),
		ConfigValue: value,
	}, &resp)
	return resp.Diagnostics.HasError()
}

func TestValidCIDR(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{ // value -> wantErr
		"10.0.0.0/24":   false,
		"2001:db8::/32": false,
		"":              false, // empty skipped
		"10.0.0.0":      true,  // missing mask
		"10.0.0.0/33":   true,  // bad mask
		"not-a-cidr":    true,
	}
	v := validCIDR()
	for value, wantErr := range cases {
		if got := runStringValidator(v, types.StringValue(value)); got != wantErr {
			t.Errorf("validCIDR(%q): wantErr=%v got=%v", value, wantErr, got)
		}
	}
	if runStringValidator(v, types.StringNull()) {
		t.Error("validCIDR(null) should not error")
	}
}

func TestValidPortRange(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{ // value -> wantErr
		"22":        false,
		"8000-9000": false,
		"0":         false,
		"65535":     false,
		"":          false, // empty skipped
		"70000":     true,  // out of range
		"-1":        true,  // empty low endpoint
		"9000-8000": true,  // low > high
		"1-2-3":     true,  // too many parts
		"abc":       true,
	}
	v := validPortRange()
	for value, wantErr := range cases {
		if got := runStringValidator(v, types.StringValue(value)); got != wantErr {
			t.Errorf("validPortRange(%q): wantErr=%v got=%v", value, wantErr, got)
		}
	}
}
