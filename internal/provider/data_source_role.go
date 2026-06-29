package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &roleDataSource{}
var _ datasource.DataSourceWithConfigure = &roleDataSource{}

type roleDataSource struct {
	dataSourceBase
}

type roleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	RoleType    types.String `tfsdk:"role_type"`
}

func NewRoleDataSource() datasource.DataSource { return &roleDataSource{} }

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud role by name (authority).",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the role."},
			"name":        dschema.StringAttribute{Required: true, Description: "Name (authority) of the role to look up."},
			"description": dschema.StringAttribute{Computed: true, Description: "Description of the role."},
			"role_type":   dschema.StringAttribute{Computed: true, Description: "Role type."},
		},
	}
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data roleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	role, err := d.client.GetRoleByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Role Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(role.ID, 10))
	data.Name = types.StringValue(role.Authority)
	data.Description = optionalString(role.Description)
	data.RoleType = optionalString(role.RoleType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
