package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &userGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &userGroupDataSource{}

type userGroupDataSource struct {
	dataSourceBase
}

type userGroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ServerGroup types.String `tfsdk:"server_group"`
}

func NewUserGroupDataSource() datasource.DataSource { return &userGroupDataSource{} }

func (d *userGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (d *userGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud user group by name.",
		Attributes: map[string]dschema.Attribute{
			"id":           dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the user group."},
			"name":         dschema.StringAttribute{Required: true, Description: "Name of the user group to look up."},
			"description":  dschema.StringAttribute{Computed: true, Description: "Description of the user group."},
			"server_group": dschema.StringAttribute{Computed: true, Description: "Linux server group name applied to members."},
		},
	}
}

func (d *userGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := d.client.GetUserGroupByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud User Group Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(group.ID, 10))
	data.Name = types.StringValue(group.Name)
	data.Description = optionalString(group.Description)
	data.ServerGroup = optionalString(group.ServerGroup)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
