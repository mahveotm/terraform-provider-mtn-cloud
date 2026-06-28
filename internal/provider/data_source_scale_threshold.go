package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &scaleThresholdDataSource{}
var _ datasource.DataSourceWithConfigure = &scaleThresholdDataSource{}

type scaleThresholdDataSource struct {
	dataSourceBase
}

type scaleThresholdDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	MinCount types.Int64  `tfsdk:"min_count"`
	MaxCount types.Int64  `tfsdk:"max_count"`
}

func NewScaleThresholdDataSource() datasource.DataSource {
	return &scaleThresholdDataSource{}
}

func (d *scaleThresholdDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scale_threshold"
}

func (d *scaleThresholdDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud autoscale threshold by name.",
		Attributes: map[string]dschema.Attribute{
			"id":        dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the scale threshold."},
			"name":      dschema.StringAttribute{Required: true, Description: "Name of the scale threshold to look up."},
			"min_count": dschema.Int64Attribute{Computed: true, Description: "Minimum number of instances."},
			"max_count": dschema.Int64Attribute{Computed: true, Description: "Maximum number of instances."},
		},
	}
}

func (d *scaleThresholdDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data scaleThresholdDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	st, err := d.client.GetScaleThresholdByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Scale Threshold Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(st.ID, 10))
	data.Name = types.StringValue(st.Name)
	data.MinCount = maybeInt64(st.MinCount)
	data.MaxCount = maybeInt64(st.MaxCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
