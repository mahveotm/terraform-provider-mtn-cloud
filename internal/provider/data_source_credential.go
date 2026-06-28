package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &credentialDataSource{}
var _ datasource.DataSourceWithConfigure = &credentialDataSource{}

type credentialDataSource struct {
	dataSourceBase
}

type credentialDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

func NewCredentialDataSource() datasource.DataSource {
	return &credentialDataSource{}
}

func (d *credentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (d *credentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud credential by name. Secret material is never returned.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the credential."},
			"name":        dschema.StringAttribute{Required: true, Description: "Name of the credential to look up."},
			"type":        dschema.StringAttribute{Computed: true, Description: "Credential type code."},
			"description": dschema.StringAttribute{Computed: true, Description: "Description of the credential."},
			"enabled":     dschema.BoolAttribute{Computed: true, Description: "Whether the credential is enabled."},
		},
	}
}

func (d *credentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data credentialDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cred, err := d.client.GetCredentialByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Credential Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(cred.ID, 10))
	data.Name = types.StringValue(cred.Name)
	data.Type = optionalString(cred.Type.Code)
	data.Description = optionalString(cred.Description)
	data.Enabled = maybeBool(cred.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
