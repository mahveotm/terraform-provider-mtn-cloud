package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtn-cloud/internal/client"
)

var _ resource.Resource = &securityGroupRuleResource{}
var _ resource.ResourceWithConfigure = &securityGroupRuleResource{}
var _ resource.ResourceWithImportState = &securityGroupRuleResource{}

type securityGroupRuleResource struct {
	client *client.Client
}

type securityGroupRuleResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	SecurityGroupID      types.String `tfsdk:"security_group_id"`
	Name                 types.String `tfsdk:"name"`
	Direction            types.String `tfsdk:"direction"`
	Policy               types.String `tfsdk:"policy"`
	Protocol             types.String `tfsdk:"protocol"`
	PortRange            types.String `tfsdk:"port_range"`
	DestinationPortRange types.String `tfsdk:"destination_port_range"`
	SourceType           types.String `tfsdk:"source_type"`
	Source               types.String `tfsdk:"source"`
	DestinationType      types.String `tfsdk:"destination_type"`
	Destination          types.String `tfsdk:"destination"`
	Ethertype            types.String `tfsdk:"ethertype"`
	Priority             types.Int64  `tfsdk:"priority"`
	Enabled              types.Bool   `tfsdk:"enabled"`
}

func NewSecurityGroupRuleResource() resource.Resource {
	return &securityGroupRuleResource{}
}

func (r *securityGroupRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_rule"
}

func (r *securityGroupRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a rule in an MTN Cloud security group.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{Computed: true},
			"security_group_id": rschema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": rschema.StringAttribute{Optional: true},
			"direction": rschema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("ingress"),
				Validators: []validator.String{stringvalidator.OneOf("ingress", "egress")},
			},
			"policy": rschema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("accept"),
				Validators: []validator.String{stringvalidator.OneOf("accept", "deny")},
			},
			"protocol": rschema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("tcp", "udp", "icmp", "any")},
			},
			"port_range": rschema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{validPortRange()},
			},
			"destination_port_range": rschema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{validPortRange()},
			},
			"source_type": rschema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("all"),
				Validators: []validator.String{stringvalidator.OneOf("cidr", "group", "instance", "all")},
			},
			"source": rschema.StringAttribute{Optional: true},
			"destination_type": rschema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("instance"),
				Validators: []validator.String{stringvalidator.OneOf("cidr", "group", "instance", "all")},
			},
			"destination": rschema.StringAttribute{Optional: true},
			"ethertype": rschema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("IPv4", "IPv6")},
			},
			"priority": rschema.Int64Attribute{
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(0)},
			},
			"enabled": rschema.BoolAttribute{Optional: true},
		},
	}
}

func (r *securityGroupRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *securityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	setRuleDefaults(&plan)
	sgID, err := strconv.ParseInt(plan.SecurityGroupID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group ID", err.Error())
		return
	}
	rule, err := r.client.CreateSecurityGroupRule(ctx, sgID, ruleInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Create MTN Cloud Security Group Rule Failed", err.Error())
		return
	}
	setRuleState(&plan, rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sgID, ruleID, err := parseRuleIDs(state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group Rule ID", err.Error())
		return
	}
	rule, err := r.client.GetSecurityGroupRule(ctx, sgID, ruleID)
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Security Group Rule Failed", err.Error())
		return
	}
	setRuleState(&state, rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityGroupRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	setRuleDefaults(&plan)
	sgID, ruleID, err := parseRuleIDs(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group Rule ID", err.Error())
		return
	}
	rule, err := r.client.UpdateSecurityGroupRule(ctx, sgID, ruleID, ruleInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Update MTN Cloud Security Group Rule Failed", err.Error())
		return
	}
	setRuleState(&plan, rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	sgID, ruleID, err := parseRuleIDs(state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Security Group Rule ID", err.Error())
		return
	}
	if err := r.client.DeleteSecurityGroupRule(ctx, sgID, ruleID); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete MTN Cloud Security Group Rule Failed", err.Error())
	}
}

func (r *securityGroupRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sgID, ruleID, err := client.ParseRuleImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("security_group_id"), strconv.FormatInt(sgID, 10))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strconv.FormatInt(ruleID, 10))...)
}

func parseRuleIDs(data securityGroupRuleResourceModel) (int64, int64, error) {
	sgID, err := strconv.ParseInt(data.SecurityGroupID.ValueString(), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid security_group_id: %w", err)
	}
	ruleID, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid id: %w", err)
	}
	return sgID, ruleID, nil
}

func setRuleDefaults(data *securityGroupRuleResourceModel) {
	if data.Direction.IsNull() || data.Direction.IsUnknown() {
		data.Direction = types.StringValue("ingress")
	}
	if data.Policy.IsNull() || data.Policy.IsUnknown() {
		data.Policy = types.StringValue("accept")
	}
	if data.SourceType.IsNull() || data.SourceType.IsUnknown() {
		data.SourceType = types.StringValue("all")
	}
	if data.DestinationType.IsNull() || data.DestinationType.IsUnknown() {
		data.DestinationType = types.StringValue("instance")
	}
}

func ruleInput(data securityGroupRuleResourceModel) client.SecurityGroupRuleInput {
	return client.SecurityGroupRuleInput{
		Name:                 data.Name.ValueString(),
		Direction:            data.Direction.ValueString(),
		Policy:               data.Policy.ValueString(),
		Protocol:             data.Protocol.ValueString(),
		PortRange:            data.PortRange.ValueString(),
		DestinationPortRange: data.DestinationPortRange.ValueString(),
		SourceType:           data.SourceType.ValueString(),
		Source:               data.Source.ValueString(),
		DestinationType:      data.DestinationType.ValueString(),
		Destination:          data.Destination.ValueString(),
		Ethertype:            data.Ethertype.ValueString(),
		Priority:             int64Ptr(data.Priority),
		Enabled:              boolPtr(data.Enabled),
	}
}

func setRuleState(data *securityGroupRuleResourceModel, rule *client.SecurityGroupRule) {
	data.ID = types.StringValue(strconv.FormatInt(rule.ID, 10))
	data.Name = optionalString(rule.Name)
	data.Direction = types.StringValue(rule.Direction)
	data.Policy = types.StringValue(rule.Policy)
	data.Protocol = optionalString(rule.Protocol)
	data.PortRange = optionalString(rule.PortRange)
	data.DestinationPortRange = optionalString(rule.DestinationPortRange)
	data.SourceType = types.StringValue(rule.SourceType)
	data.Source = optionalString(rule.Source)
	data.DestinationType = types.StringValue(rule.DestinationType)
	data.Destination = optionalString(rule.Destination)
	data.Ethertype = optionalString(rule.Ethertype)
	data.Priority = maybeInt64(rule.Priority)
	data.Enabled = maybeBool(rule.Enabled)
}

func pathRootID() path.Path {
	return path.Root("id")
}
