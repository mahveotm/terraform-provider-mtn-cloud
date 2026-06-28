package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

var _ datasource.DataSource = &servicePlanDataSource{}
var _ datasource.DataSourceWithConfigure = &servicePlanDataSource{}

type servicePlanDataSource struct {
	client *client.Client
}

type servicePlanDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Group      types.String `tfsdk:"group"`
	Type       types.String `tfsdk:"type"`
	Code       types.String `tfsdk:"code"`
	MaxCPU     types.Int64  `tfsdk:"max_cpu"`
	MaxMemory  types.Int64  `tfsdk:"max_memory"`
	MaxStorage types.Int64  `tfsdk:"max_storage"`
}

func NewServicePlanDataSource() datasource.DataSource {
	return &servicePlanDataSource{}
}

func (d *servicePlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_plan"
}

func (d *servicePlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud service plan by name/code for a group and instance type.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the service plan."},
			"name":        dschema.StringAttribute{Required: true, Description: "Name of the service plan to look up (e.g. `G2S4`)."},
			"group":       dschema.StringAttribute{Required: true, Description: "Name of the group the plan is available in."},
			"type":        dschema.StringAttribute{Required: true, Description: "Instance type code the plan applies to (e.g. `MTN-CS10`)."},
			"code":        dschema.StringAttribute{Computed: true, Description: "Code of the service plan."},
			"max_cpu":     dschema.Int64Attribute{Computed: true, Description: "Maximum number of vCPUs the plan provides."},
			"max_memory":  dschema.Int64Attribute{Computed: true, Description: "Maximum memory the plan provides, in bytes."},
			"max_storage": dschema.Int64Attribute{Computed: true, Description: "Maximum storage the plan provides, in bytes."},
		},
	}
}

func (d *servicePlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := configuredClient(req.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client.")
		return
	}
	d.client = apiClient
}

func (d *servicePlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data servicePlanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, instanceType, err := resolveGroupAndInstanceType(ctx, d.client, data.Group.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Resolve MTN Cloud Plan Context Failed", err.Error())
		return
	}
	if len(group.CloudIDs) == 0 || instanceType.DefaultLayoutID == nil {
		resp.Diagnostics.AddError("Invalid MTN Cloud Plan Context", "Group must have a cloud ID and instance type must have a default layout ID.")
		return
	}

	plan, err := d.client.GetServicePlan(ctx, data.Name.ValueString(), group.CloudIDs[0], *instanceType.DefaultLayoutID, group.ID)
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Service Plan Failed", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(plan.ID, 10))
	data.Name = types.StringValue(plan.Name)
	data.Code = types.StringValue(plan.Code)
	data.MaxCPU = maybeInt64(plan.MaxCPU)
	data.MaxMemory = maybeInt64(plan.MaxMemory)
	data.MaxStorage = maybeInt64(plan.MaxStorage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func resolveGroupAndInstanceType(ctx context.Context, apiClient *client.Client, groupName, typeCode string) (*client.Group, *client.InstanceType, error) {
	group, err := apiClient.GetGroupByName(ctx, groupName)
	if err != nil {
		return nil, nil, err
	}
	instanceType, err := apiClient.GetInstanceTypeByCode(ctx, typeCode)
	if err != nil {
		return nil, nil, err
	}
	return group, instanceType, nil
}
