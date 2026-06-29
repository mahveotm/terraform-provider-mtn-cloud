package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

var _ resource.Resource = &roleResource{}
var _ resource.ResourceWithConfigure = &roleResource{}
var _ resource.ResourceWithImportState = &roleResource{}

type roleResource struct {
	resourceBase
}

type roleResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	RoleType          types.String `tfsdk:"role_type"`
	Multitenant       types.Bool   `tfsdk:"multitenant"`
	MultitenantLocked types.Bool   `tfsdk:"multitenant_locked"`
	PermissionSet     types.String `tfsdk:"permission_set"`
}

func NewRoleResource() resource.Resource { return &roleResource{} }

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud role. `permission_set` is a JSON document in the Morpheus role " +
			"API shape (global access defaults plus per-feature/cloud/instance-type/task/blueprint/… permission " +
			"lists); it is config-authoritative and not read back. Note: creating `account` (tenant) roles needs " +
			"admin-accounts access, which the customer-admin token typically lacks.",
		Attributes: map[string]rschema.Attribute{
			"id":          computedIDAttribute("Numeric identifier of the role."),
			"name":        rschema.StringAttribute{Required: true, Description: "Name (authority) of the role."},
			"description": rschema.StringAttribute{Optional: true, Computed: true, Description: "Description of the role."},
			"role_type": rschema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("user"),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.OneOf(client.RoleTypes...)},
				Description:   "Role type: " + joinQuoted(client.RoleTypes) + " (default `user`). Changing it forces a new role.",
			},
			"multitenant": rschema.BoolAttribute{
				Optional: true, Computed: true,
				Description: "Whether the role applies across all tenants (master-tenant roles).",
			},
			"multitenant_locked": rschema.BoolAttribute{
				Optional: true, Computed: true,
				Description: "Whether sub-tenants are prevented from modifying the role.",
			},
			"permission_set": rschema.StringAttribute{
				Optional: true,
				Description: "JSON document of role permissions in the Morpheus role API shape, e.g. " +
					"`jsonencode({ globalSiteAccess = \"all\", featurePermissions = [{ code = \"admin-users\", access = \"full\" }] })`. " +
					"Config-authoritative (not read back).",
			},
		},
	}
}

func (r *roleResource) input(plan roleResourceModel) client.RoleInput {
	return client.RoleInput{
		Name:              plan.Name.ValueString(),
		Description:       plan.Description.ValueString(),
		RoleType:          plan.RoleType.ValueString(),
		Multitenant:       boolPtr(plan.Multitenant),
		MultitenantLocked: boolPtr(plan.MultitenantLocked),
		PermissionSet:     plan.PermissionSet.ValueString(),
	}
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := r.client.CreateRole(ctx, r.input(plan))
	if err != nil {
		opError(&resp.Diagnostics, "Create", "Role", err)
		return
	}
	setRoleState(&plan, role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "Role", &resp.Diagnostics)
	if !ok {
		return
	}
	role, err := r.client.GetRole(ctx, id)
	if handleReadError(ctx, err, "Role", &resp.State, &resp.Diagnostics) {
		return
	}
	setRoleState(&state, role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(plan.ID, "Role", &resp.Diagnostics)
	if !ok {
		return
	}
	role, err := r.client.UpdateRole(ctx, id, r.input(plan))
	if err != nil {
		opError(&resp.Diagnostics, "Update", "Role", err)
		return
	}
	setRoleState(&plan, role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "Role", &resp.Diagnostics)
	if !ok {
		return
	}
	if err := r.client.DeleteRole(ctx, id); err != nil && !client.IsNotFound(err) {
		opError(&resp.Diagnostics, "Delete", "Role", err)
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

// setRoleState reconciles stable role metadata; permission_set is
// config-authoritative and kept from prior state.
func setRoleState(data *roleResourceModel, role *client.Role) {
	data.ID = types.StringValue(strconv.FormatInt(role.ID, 10))
	data.Name = types.StringValue(role.Authority)
	data.Description = mergeAPIString(data.Description, role.Description)
	data.RoleType = mergeAPIString(data.RoleType, role.RoleType)
	data.Multitenant = mergeAPIBool(data.Multitenant, role.Multitenant)
	data.MultitenantLocked = mergeAPIBool(data.MultitenantLocked, role.MultitenantLocked)
}
