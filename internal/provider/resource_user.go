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

var _ resource.Resource = &userResource{}
var _ resource.ResourceWithConfigure = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

type userResource struct {
	resourceBase
}

type userResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Username             types.String `tfsdk:"username"`
	Email                types.String `tfsdk:"email"`
	Password             types.String `tfsdk:"password"`
	FirstName            types.String `tfsdk:"first_name"`
	LastName             types.String `tfsdk:"last_name"`
	RoleIDs              types.List   `tfsdk:"role_ids"`
	PasswordExpired      types.Bool   `tfsdk:"password_expired"`
	ReceiveNotifications types.Bool   `tfsdk:"receive_notifications"`
	LinuxUsername        types.String `tfsdk:"linux_username"`
	LinuxPassword        types.String `tfsdk:"linux_password"`
	LinuxKeyPairID       types.Int64  `tfsdk:"linux_keypair_id"`
	WindowsUsername      types.String `tfsdk:"windows_username"`
	WindowsPassword      types.String `tfsdk:"windows_password"`
}

func NewUserResource() resource.Resource { return &userResource{} }

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud user in the caller's tenant. `password` and the Linux/Windows " +
			"passwords are write-only (never returned by the API). Cross-tenant placement is not supported " +
			"(it needs admin-accounts access the customer-admin token lacks).",
		Attributes: map[string]rschema.Attribute{
			"id":         computedIDAttribute("Numeric identifier of the user."),
			"username":   rschema.StringAttribute{Required: true, Description: "Username (login) of the user."},
			"email":      rschema.StringAttribute{Required: true, Description: "Email address of the user."},
			"password":   rschema.StringAttribute{Required: true, Sensitive: true, Description: "Password (write-only; never returned by the API)."},
			"first_name": rschema.StringAttribute{Optional: true, Computed: true, Description: "First name."},
			"last_name":  rschema.StringAttribute{Optional: true, Computed: true, Description: "Last name."},
			"role_ids": rschema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "IDs of the roles assigned to the user.",
			},
			"password_expired":      rschema.BoolAttribute{Optional: true, Description: "Whether the password is marked expired (forces a reset at next login). Write-only directive."},
			"receive_notifications": rschema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the user receives notification emails."},
			"linux_username":        rschema.StringAttribute{Optional: true, Description: "Default Linux username for provisioned instances."},
			"linux_password":        rschema.StringAttribute{Optional: true, Sensitive: true, Description: "Default Linux password (write-only)."},
			"linux_keypair_id":      rschema.Int64Attribute{Optional: true, Description: "Default Linux key pair ID."},
			"windows_username":      rschema.StringAttribute{Optional: true, Description: "Default Windows username for provisioned instances."},
			"windows_password":      rschema.StringAttribute{Optional: true, Sensitive: true, Description: "Default Windows password (write-only)."},
		},
	}
}

func (r *userResource) input(ctx context.Context, plan userResourceModel) client.UserInput {
	return client.UserInput{
		Username:             plan.Username.ValueString(),
		Email:                plan.Email.ValueString(),
		Password:             plan.Password.ValueString(),
		FirstName:            plan.FirstName.ValueString(),
		LastName:             plan.LastName.ValueString(),
		RoleIDs:              int64Slice(ctx, plan.RoleIDs),
		PasswordExpired:      boolPtr(plan.PasswordExpired),
		ReceiveNotifications: boolPtr(plan.ReceiveNotifications),
		LinuxUsername:        plan.LinuxUsername.ValueString(),
		LinuxPassword:        plan.LinuxPassword.ValueString(),
		LinuxKeyPairID:       int64Ptr(plan.LinuxKeyPairID),
		WindowsUsername:      plan.WindowsUsername.ValueString(),
		WindowsPassword:      plan.WindowsPassword.ValueString(),
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := r.client.CreateUser(ctx, r.input(ctx, plan))
	if err != nil {
		opError(&resp.Diagnostics, "Create", "User", err)
		return
	}
	resp.Diagnostics.Append(setUserState(ctx, &plan, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "User", &resp.Diagnostics)
	if !ok {
		return
	}
	user, err := r.client.GetUser(ctx, id)
	if handleReadError(ctx, err, "User", &resp.State, &resp.Diagnostics) {
		return
	}
	resp.Diagnostics.Append(setUserState(ctx, &state, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(plan.ID, "User", &resp.Diagnostics)
	if !ok {
		return
	}
	user, err := r.client.UpdateUser(ctx, id, r.input(ctx, plan))
	if err != nil {
		opError(&resp.Diagnostics, "Update", "User", err)
		return
	}
	resp.Diagnostics.Append(setUserState(ctx, &plan, user)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "User", &resp.Diagnostics)
	if !ok {
		return
	}
	if err := r.client.DeleteUser(ctx, id); err != nil && !client.IsNotFound(err) {
		opError(&resp.Diagnostics, "Delete", "User", err)
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

// setUserState reconciles non-secret metadata; passwords are write-only and kept
// from prior state. role_ids round-trips from the API (sorted) for import.
func setUserState(ctx context.Context, data *userResourceModel, user *client.User) diag.Diagnostics {
	data.ID = types.StringValue(strconv.FormatInt(user.ID, 10))
	data.Username = types.StringValue(user.Username)
	data.Email = types.StringValue(user.Email)
	data.FirstName = mergeAPIString(data.FirstName, user.FirstName)
	data.LastName = mergeAPIString(data.LastName, user.LastName)
	data.ReceiveNotifications = mergeAPIBool(data.ReceiveNotifications, user.ReceiveNotifications)
	roleIDs, diags := types.ListValueFrom(ctx, types.Int64Type, user.RoleIDList())
	data.RoleIDs = roleIDs
	return diags
}
