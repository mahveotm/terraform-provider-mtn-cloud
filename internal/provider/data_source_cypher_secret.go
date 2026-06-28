package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &cypherSecretDataSource{}
var _ datasource.DataSourceWithConfigure = &cypherSecretDataSource{}

type cypherSecretDataSource struct {
	dataSourceBase
}

type cypherSecretDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
	TTL   types.Int64  `tfsdk:"ttl"`
}

func NewCypherSecretDataSource() datasource.DataSource {
	return &cypherSecretDataSource{}
}

func (d *cypherSecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cypher_secret"
}

func (d *cypherSecretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Reads a secret from the MTN Cloud Cypher store (under the `secret/` mount) by key.",
		Attributes: map[string]dschema.Attribute{
			"id":    dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the cypher entry."},
			"key":   dschema.StringAttribute{Required: true, Description: "The secret path under the `secret/` mount to read."},
			"value": dschema.StringAttribute{Computed: true, Sensitive: true, Description: "The decrypted secret value."},
			"ttl":   dschema.Int64Attribute{Computed: true, Description: "Remaining lease time-to-live in seconds. 0 means no expiry."},
		},
	}
}

func (d *cypherSecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cypherSecretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.GetCypher(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Cypher Secret Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(result.Cypher.ID, 10))
	data.Value = types.StringValue(result.Value)
	data.TTL = maybeInt64(result.LeaseDuration)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
