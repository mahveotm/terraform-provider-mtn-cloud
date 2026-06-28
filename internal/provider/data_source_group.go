package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtn-cloud/internal/client"
)

var _ datasource.DataSource = &groupDataSource{}
var _ datasource.DataSourceWithConfigure = &groupDataSource{}

type groupDataSource struct {
	client *client.Client
}

type groupDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	CloudIDs types.List   `tfsdk:"cloud_ids"`
	Location types.String `tfsdk:"location"`
	Active   types.Bool   `tfsdk:"active"`
}

func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud group/site by name.",
		Attributes: map[string]dschema.Attribute{
			"id":   dschema.StringAttribute{Computed: true},
			"name": dschema.StringAttribute{Required: true},
			"cloud_ids": dschema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"location": dschema.StringAttribute{Computed: true},
			"active":   dschema.BoolAttribute{Computed: true},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := d.client.GetGroupByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Group Failed", err.Error())
		return
	}

	cloudIDs, diags := types.ListValueFrom(ctx, types.Int64Type, group.CloudIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(group.ID, 10))
	data.Name = types.StringValue(group.Name)
	data.CloudIDs = cloudIDs
	data.Location = types.StringValue(group.Location)
	data.Active = maybeBool(group.Active)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
