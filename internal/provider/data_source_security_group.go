package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtn-cloud/internal/client"
)

var _ datasource.DataSource = &securityGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &securityGroupDataSource{}

type securityGroupDataSource struct {
	client *client.Client
}

type securityGroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Active      types.Bool   `tfsdk:"active"`
}

func NewSecurityGroupDataSource() datasource.DataSource {
	return &securityGroupDataSource{}
}

func (d *securityGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (d *securityGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud security group by name.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Computed: true},
			"name":        dschema.StringAttribute{Required: true},
			"description": dschema.StringAttribute{Computed: true},
			"active":      dschema.BoolAttribute{Computed: true},
		},
	}
}

func (d *securityGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data securityGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sg, err := d.client.GetSecurityGroupByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Security Group Failed", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(sg.ID, 10))
	data.Description = optionalString(sg.Description)
	data.Active = maybeBool(sg.Active)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
