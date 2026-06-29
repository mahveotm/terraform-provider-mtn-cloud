package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

var _ resource.Resource = &userGroupResource{}
var _ resource.ResourceWithConfigure = &userGroupResource{}
var _ resource.ResourceWithImportState = &userGroupResource{}

type userGroupResource struct {
	resourceBase
}

type userGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	SudoAccess  types.Bool   `tfsdk:"sudo_access"`
	ServerGroup types.String `tfsdk:"server_group"`
	UserIDs     types.List   `tfsdk:"user_ids"`
}

func NewUserGroupResource() resource.Resource { return &userGroupResource{} }

func (r *userGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (r *userGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud user group (a set of users with a shared Linux login and sudo policy).",
		Attributes: map[string]rschema.Attribute{
			"id":          computedIDAttribute("Numeric identifier of the user group."),
			"name":        rschema.StringAttribute{Required: true, Description: "Name of the user group."},
			"description": rschema.StringAttribute{Optional: true, Computed: true, Description: "Description of the user group."},
			"sudo_access": rschema.BoolAttribute{Optional: true, Computed: true, Description: "Whether members get sudo access on provisioned instances."},
			"server_group": rschema.StringAttribute{
				Optional: true, Computed: true,
				Description: "Linux server group name applied to members.",
			},
			"user_ids": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "IDs of the users that belong to the group.",
			},
		},
	}
}

func (r *userGroupResource) input(ctx context.Context, plan userGroupResourceModel) client.UserGroupInput {
	return client.UserGroupInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		SudoAccess:  boolPtr(plan.SudoAccess),
		ServerGroup: plan.ServerGroup.ValueString(),
		UserIDs:     int64Slice(ctx, plan.UserIDs),
	}
}

func (r *userGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.CreateUserGroup(ctx, r.input(ctx, plan))
	if err != nil {
		opError(&resp.Diagnostics, "Create", "User Group", err)
		return
	}
	resp.Diagnostics.Append(setUserGroupState(ctx, &plan, group)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "User Group", &resp.Diagnostics)
	if !ok {
		return
	}
	group, err := r.client.GetUserGroup(ctx, id)
	if handleReadError(ctx, err, "User Group", &resp.State, &resp.Diagnostics) {
		return
	}
	resp.Diagnostics.Append(setUserGroupState(ctx, &state, group)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(plan.ID, "User Group", &resp.Diagnostics)
	if !ok {
		return
	}
	group, err := r.client.UpdateUserGroup(ctx, id, r.input(ctx, plan))
	if err != nil {
		opError(&resp.Diagnostics, "Update", "User Group", err)
		return
	}
	resp.Diagnostics.Append(setUserGroupState(ctx, &plan, group)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "User Group", &resp.Diagnostics)
	if !ok {
		return
	}
	if err := r.client.DeleteUserGroup(ctx, id); err != nil && !client.IsNotFound(err) {
		opError(&resp.Diagnostics, "Delete", "User Group", err)
	}
}

func (r *userGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

func setUserGroupState(ctx context.Context, data *userGroupResourceModel, group *client.UserGroup) diag.Diagnostics {
	data.ID = types.StringValue(strconv.FormatInt(group.ID, 10))
	data.Name = types.StringValue(group.Name)
	data.Description = mergeAPIString(data.Description, group.Description)
	data.ServerGroup = mergeAPIString(data.ServerGroup, group.ServerGroup)
	data.SudoAccess = mergeAPIBool(data.SudoAccess, group.SudoUser)
	userIDs, diags := types.ListValueFrom(ctx, types.Int64Type, group.UserIDList())
	data.UserIDs = userIDs
	return diags
}
