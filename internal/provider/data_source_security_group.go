package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &securityGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &securityGroupDataSource{}

type securityGroupDataSource struct {
	dataSourceBase
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
			"id":          dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the security group."},
			"name":        dschema.StringAttribute{Required: true, Description: "Name of the security group to look up."},
			"description": dschema.StringAttribute{Computed: true, Description: "Description of the security group."},
			"active":      dschema.BoolAttribute{Computed: true, Description: "Whether the security group is active."},
		},
	}
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
