package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &budgetDataSource{}
var _ datasource.DataSourceWithConfigure = &budgetDataSource{}

type budgetDataSource struct {
	dataSourceBase
}

type budgetDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Interval types.String `tfsdk:"interval"`
	Year     types.String `tfsdk:"year"`
	Currency types.String `tfsdk:"currency"`
}

func NewBudgetDataSource() datasource.DataSource {
	return &budgetDataSource{}
}

func (d *budgetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_budget"
}

func (d *budgetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud budget by name.",
		Attributes: map[string]dschema.Attribute{
			"id":       dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the budget."},
			"name":     dschema.StringAttribute{Required: true, Description: "Name of the budget to look up."},
			"enabled":  dschema.BoolAttribute{Computed: true, Description: "Whether the budget is enabled."},
			"interval": dschema.StringAttribute{Computed: true, Description: "Budget interval."},
			"year":     dschema.StringAttribute{Computed: true, Description: "Calendar year of the budget."},
			"currency": dschema.StringAttribute{Computed: true, Description: "Currency code of the budget."},
		},
	}
}

func (d *budgetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data budgetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	budget, err := d.client.GetBudgetByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Budget Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(budget.ID, 10))
	data.Name = types.StringValue(budget.Name)
	data.Enabled = maybeBool(budget.Enabled)
	data.Interval = optionalString(budget.Interval)
	data.Year = optionalString(budget.Year)
	data.Currency = optionalString(budget.Currency)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
