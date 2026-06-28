package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &keyPairDataSource{}
var _ datasource.DataSourceWithConfigure = &keyPairDataSource{}

type keyPairDataSource struct {
	dataSourceBase
}

type keyPairDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	PublicKey types.String `tfsdk:"public_key"`
}

func NewKeyPairDataSource() datasource.DataSource {
	return &keyPairDataSource{}
}

func (d *keyPairDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_pair"
}

func (d *keyPairDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud SSH key pair by name.",
		Attributes: map[string]dschema.Attribute{
			"id":         dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the key pair."},
			"name":       dschema.StringAttribute{Required: true, Description: "Name of the key pair to look up."},
			"public_key": dschema.StringAttribute{Computed: true, Description: "The public key material."},
		},
	}
}

func (d *keyPairDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data keyPairDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	kp, err := d.client.GetKeyPairByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Key Pair Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(kp.ID, 10))
	data.Name = types.StringValue(kp.Name)
	data.PublicKey = optionalString(kp.PublicKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
