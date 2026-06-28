package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

// parseID converts a string state ID into the numeric ID the API uses. On failure
// it records a diagnostic and returns ok=false so the caller can bail out.
func parseID(id types.String, label string, diags *diag.Diagnostics) (int64, bool) {
	parsed, err := strconv.ParseInt(id.ValueString(), 10, 64)
	if err != nil {
		diags.AddError(fmt.Sprintf("Invalid %s ID", label),
			fmt.Sprintf("Could not parse %q as a numeric ID: %s", id.ValueString(), err))
		return 0, false
	}
	return parsed, true
}

// handleReadError standardizes Read-path error handling: a 404 removes the
// resource from state (drift), any other error is surfaced. It returns stop=true
// when the caller should return without setting state.
func handleReadError(ctx context.Context, err error, label string, state *tfsdk.State, diags *diag.Diagnostics) (stop bool) {
	if err == nil {
		return false
	}
	if client.IsNotFound(err) {
		state.RemoveResource(ctx)
		return true
	}
	opError(diags, "Read", label, err)
	return true
}

// opError records a consistent diagnostic for a failed CRUD operation, e.g.
// "Create MTN Cloud Credential Failed".
func opError(diags *diag.Diagnostics, op, label string, err error) {
	diags.AddError(fmt.Sprintf("%s MTN Cloud %s Failed", op, label), err.Error())
}
