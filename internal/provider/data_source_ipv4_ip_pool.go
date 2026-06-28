package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ipPoolDataSource{}
var _ datasource.DataSourceWithConfigure = &ipPoolDataSource{}

type ipPoolDataSource struct {
	dataSourceBase
}

type ipPoolDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Gateway types.String `tfsdk:"gateway"`
	Netmask types.String `tfsdk:"netmask"`
}

func NewIPPoolDataSource() datasource.DataSource {
	return &ipPoolDataSource{}
}

func (d *ipPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipv4_ip_pool"
}

func (d *ipPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud IPv4 address pool by name.",
		Attributes: map[string]dschema.Attribute{
			"id":      dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the IP pool."},
			"name":    dschema.StringAttribute{Required: true, Description: "Name of the IP pool to look up."},
			"gateway": dschema.StringAttribute{Computed: true, Description: "Gateway IP address for the pool."},
			"netmask": dschema.StringAttribute{Computed: true, Description: "Netmask for the pool."},
		},
	}
}

func (d *ipPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ipPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool, err := d.client.GetIPPoolByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud IP Pool Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(pool.ID, 10))
	data.Name = types.StringValue(pool.Name)
	data.Gateway = optionalString(pool.Gateway)
	data.Netmask = optionalString(pool.Netmask)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
