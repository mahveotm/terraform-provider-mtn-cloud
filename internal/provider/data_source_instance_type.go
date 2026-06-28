package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtn-cloud/internal/client"
)

var _ datasource.DataSource = &instanceTypeDataSource{}
var _ datasource.DataSourceWithConfigure = &instanceTypeDataSource{}

type instanceTypeDataSource struct {
	client *client.Client
}

type instanceTypeDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Code            types.String `tfsdk:"code"`
	DefaultLayoutID types.Int64  `tfsdk:"default_layout_id"`
}

func NewInstanceTypeDataSource() datasource.DataSource {
	return &instanceTypeDataSource{}
}

func (d *instanceTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_type"
}

func (d *instanceTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud instance type by code.",
		Attributes: map[string]dschema.Attribute{
			"id":                dschema.StringAttribute{Computed: true},
			"name":              dschema.StringAttribute{Computed: true},
			"code":              dschema.StringAttribute{Required: true},
			"default_layout_id": dschema.Int64Attribute{Computed: true},
		},
	}
}

func (d *instanceTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *instanceTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceType, err := d.client.GetInstanceTypeByCode(ctx, data.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Instance Type Failed", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(instanceType.ID, 10))
	data.Name = types.StringValue(instanceType.Name)
	data.Code = types.StringValue(instanceType.Code)
	data.DefaultLayoutID = maybeInt64(instanceType.DefaultLayoutID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
