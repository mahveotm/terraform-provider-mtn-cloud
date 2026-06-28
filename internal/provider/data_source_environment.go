package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &environmentDataSource{}
var _ datasource.DataSourceWithConfigure = &environmentDataSource{}

type environmentDataSource struct {
	dataSourceBase
}

type environmentDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Code        types.String `tfsdk:"code"`
	Visibility  types.String `tfsdk:"visibility"`
	Active      types.Bool   `tfsdk:"active"`
}

func NewEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud environment by name.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the environment."},
			"name":        dschema.StringAttribute{Required: true, Description: "Name of the environment to look up."},
			"description": dschema.StringAttribute{Computed: true, Description: "Description of the environment."},
			"code":        dschema.StringAttribute{Computed: true, Description: "Short code identifying the environment."},
			"visibility":  dschema.StringAttribute{Computed: true, Description: "Visibility of the environment."},
			"active":      dschema.BoolAttribute{Computed: true, Description: "Whether the environment is enabled."},
		},
	}
}

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data environmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	env, err := d.client.GetEnvironmentByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Environment Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(env.ID, 10))
	data.Name = types.StringValue(env.Name)
	data.Description = optionalString(env.Description)
	data.Code = optionalString(env.Code)
	data.Visibility = optionalString(env.Visibility)
	data.Active = maybeBool(env.Active)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
