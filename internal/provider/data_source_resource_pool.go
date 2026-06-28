package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &resourcePoolDataSource{}
var _ datasource.DataSourceWithConfigure = &resourcePoolDataSource{}

type resourcePoolDataSource struct {
	dataSourceBase
}

type resourcePoolDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Group types.String `tfsdk:"group"`
	Code  types.String `tfsdk:"code"`
}

func NewResourcePoolDataSource() datasource.DataSource {
	return &resourcePoolDataSource{}
}

func (d *resourcePoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_pool"
}

func (d *resourcePoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud resource pool by name/code within a group.",
		Attributes: map[string]dschema.Attribute{
			"id":    dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the resource pool."},
			"name":  dschema.StringAttribute{Required: true, Description: "Name of the resource pool to look up."},
			"group": dschema.StringAttribute{Required: true, Description: "Name of the group the resource pool belongs to."},
			"code":  dschema.StringAttribute{Computed: true, Description: "Code of the resource pool (e.g. `pool-123`)."},
		},
	}
}

func (d *resourcePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resourcePoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := d.client.GetGroupByName(ctx, data.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Resolve MTN Cloud Group Failed", err.Error())
		return
	}
	pool, err := d.client.GetResourcePool(ctx, data.Name.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Resource Pool Failed", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(pool.ID, 10))
	data.Name = types.StringValue(pool.Name)
	data.Code = types.StringValue(pool.Code)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
