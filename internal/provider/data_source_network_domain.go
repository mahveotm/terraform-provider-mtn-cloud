package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &networkDomainDataSource{}
var _ datasource.DataSourceWithConfigure = &networkDomainDataSource{}

type networkDomainDataSource struct {
	dataSourceBase
}

type networkDomainDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	FQDN        types.String `tfsdk:"fqdn"`
	Visibility  types.String `tfsdk:"visibility"`
	Active      types.Bool   `tfsdk:"active"`
}

func NewNetworkDomainDataSource() datasource.DataSource {
	return &networkDomainDataSource{}
}

func (d *networkDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_domain"
}

func (d *networkDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud network domain by name.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the network domain."},
			"name":        dschema.StringAttribute{Required: true, Description: "Name of the network domain to look up."},
			"description": dschema.StringAttribute{Computed: true, Description: "Description of the network domain."},
			"fqdn":        dschema.StringAttribute{Computed: true, Description: "Fully qualified domain name."},
			"visibility":  dschema.StringAttribute{Computed: true, Description: "Visibility of the network domain."},
			"active":      dschema.BoolAttribute{Computed: true, Description: "Whether the domain is active."},
		},
	}
}

func (d *networkDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data networkDomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	domain, err := d.client.GetNetworkDomainByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Network Domain Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(domain.ID, 10))
	data.Name = types.StringValue(domain.Name)
	data.Description = optionalString(domain.Description)
	data.FQDN = optionalString(domain.FQDN)
	data.Visibility = optionalString(domain.Visibility)
	data.Active = maybeBool(domain.Active)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
