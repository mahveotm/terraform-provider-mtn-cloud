package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &userDataSource{}
var _ datasource.DataSourceWithConfigure = &userDataSource{}

type userDataSource struct {
	dataSourceBase
}

type userDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
}

func NewUserDataSource() datasource.DataSource { return &userDataSource{} }

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud user by username.",
		Attributes: map[string]dschema.Attribute{
			"id":         dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the user."},
			"username":   dschema.StringAttribute{Required: true, Description: "Username to look up."},
			"email":      dschema.StringAttribute{Computed: true, Description: "Email address of the user."},
			"first_name": dschema.StringAttribute{Computed: true, Description: "First name."},
			"last_name":  dschema.StringAttribute{Computed: true, Description: "Last name."},
		},
	}
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := d.client.GetUserByName(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud User Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(user.ID, 10))
	data.Username = types.StringValue(user.Username)
	data.Email = optionalString(user.Email)
	data.FirstName = optionalString(user.FirstName)
	data.LastName = optionalString(user.LastName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
