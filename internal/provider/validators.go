package provider

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// cidrValidator ensures a string is a valid CIDR block (e.g. 10.0.0.0/24).
type cidrValidator struct{}

func (cidrValidator) Description(_ context.Context) string {
	return "value must be a valid CIDR block (e.g. 10.0.0.0/24)"
}

func (v cidrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (cidrValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if value == "" {
		return
	}
	if _, _, err := net.ParseCIDR(value); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid CIDR",
			fmt.Sprintf("%q is not a valid CIDR block (expected e.g. 10.0.0.0/24): %s", value, err))
	}
}

func validCIDR() validator.String { return cidrValidator{} }

// portRangeValidator ensures a string is a single port ("22") or an inclusive
// range ("8000-9000"), with each endpoint in [0, 65535] and low <= high.
type portRangeValidator struct{}

func (portRangeValidator) Description(_ context.Context) string {
	return `value must be a port ("22") or range ("8000-9000") within 0-65535`
}

func (v portRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (portRangeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := strings.TrimSpace(req.ConfigValue.ValueString())
	if value == "" {
		return
	}
	addErr := func() {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Port Range",
			fmt.Sprintf("%q must be a port like \"22\" or a range like \"8000-9000\" within 0-65535", value))
	}
	parts := strings.Split(value, "-")
	if len(parts) > 2 {
		addErr()
		return
	}
	var ports []int
	for _, part := range parts {
		port, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || port < 0 || port > 65535 {
			addErr()
			return
		}
		ports = append(ports, port)
	}
	if len(ports) == 2 && ports[0] > ports[1] {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Port Range",
			fmt.Sprintf("%q has a start port greater than its end port", value))
	}
}

func validPortRange() validator.String { return portRangeValidator{} }
