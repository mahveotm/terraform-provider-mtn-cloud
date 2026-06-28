package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &wikiPageDataSource{}
var _ datasource.DataSourceWithConfigure = &wikiPageDataSource{}

type wikiPageDataSource struct {
	dataSourceBase
}

type wikiPageDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Category types.String `tfsdk:"category"`
	Content  types.String `tfsdk:"content"`
}

func NewWikiPageDataSource() datasource.DataSource {
	return &wikiPageDataSource{}
}

func (d *wikiPageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wiki_page"
}

func (d *wikiPageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud wiki page by name.",
		Attributes: map[string]dschema.Attribute{
			"id":       dschema.StringAttribute{Computed: true, Description: "Numeric identifier of the wiki page."},
			"name":     dschema.StringAttribute{Required: true, Description: "Name (title) of the wiki page to look up."},
			"category": dschema.StringAttribute{Computed: true, Description: "Category the wiki page belongs to."},
			"content":  dschema.StringAttribute{Computed: true, Description: "Markdown content of the wiki page."},
		},
	}
}

func (d *wikiPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data wikiPageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.GetWikiPageByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Wiki Page Failed", err.Error())
		return
	}
	data.ID = types.StringValue(strconv.FormatInt(page.ID, 10))
	data.Name = types.StringValue(page.Name)
	data.Category = optionalString(page.Category)
	data.Content = optionalString(page.Content)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
