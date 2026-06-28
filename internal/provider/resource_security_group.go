package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

var _ resource.Resource = &securityGroupResource{}
var _ resource.ResourceWithConfigure = &securityGroupResource{}
var _ resource.ResourceWithImportState = &securityGroupResource{}

type securityGroupResource struct {
	client *client.Client
}

type securityGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Active      types.Bool   `tfsdk:"active"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

func (r *securityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud security group.",
		Attributes: map[string]rschema.Attribute{
			"id":          rschema.StringAttribute{Computed: true, Description: "Numeric identifier of the security group."},
			"name":        rschema.StringAttribute{Required: true, Description: "Name of the security group."},
			"description": rschema.StringAttribute{Optional: true, Description: "Human-readable description of the security group."},
			"active":      rschema.BoolAttribute{Computed: true, Description: "Whether the security group is active."},
			"enabled":     rschema.BoolAttribute{Computed: true, Description: "Whether the security group is enabled."},
		},
	}
}

func (r *securityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := configuredClient(req.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client.")
		return
	}
	r.client = apiClient
}

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sg, err := r.client.CreateSecurityGroup(ctx, plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create MTN Cloud Security Group Failed", err.Error())
		return
	}
	setSecurityGroupState(&plan, sg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group ID", err.Error())
		return
	}
	sg, err := r.client.GetSecurityGroup(ctx, id)
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Security Group Failed", err.Error())
		return
	}
	setSecurityGroupState(&state, sg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group ID", err.Error())
		return
	}
	sg, err := r.client.UpdateSecurityGroup(ctx, id, plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update MTN Cloud Security Group Failed", err.Error())
		return
	}
	setSecurityGroupState(&plan, sg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group ID", err.Error())
		return
	}
	if err := r.client.DeleteSecurityGroup(ctx, id); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete MTN Cloud Security Group Failed", err.Error())
	}
}

func (r *securityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

func setSecurityGroupState(data *securityGroupResourceModel, sg *client.SecurityGroup) {
	data.ID = types.StringValue(strconv.FormatInt(sg.ID, 10))
	data.Name = types.StringValue(sg.Name)
	data.Description = optionalString(sg.Description)
	data.Active = maybeBool(sg.Active)
	data.Enabled = maybeBool(sg.Enabled)
}
