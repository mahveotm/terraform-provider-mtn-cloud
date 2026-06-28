package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtn-cloud/internal/client"
)

var _ datasource.DataSource = &virtualImageDataSource{}
var _ datasource.DataSourceWithConfigure = &virtualImageDataSource{}

type virtualImageDataSource struct {
	client *client.Client
}

type virtualImageDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Code      types.String `tfsdk:"code"`
	ImageType types.String `tfsdk:"image_type"`
	OS        types.String `tfsdk:"os"`
	IsPublic  types.Bool   `tfsdk:"is_public"`
}

func NewVirtualImageDataSource() datasource.DataSource {
	return &virtualImageDataSource{}
}

func (d *virtualImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_image"
}

func (d *virtualImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up an MTN Cloud virtual image by name.",
		Attributes: map[string]dschema.Attribute{
			"id":         dschema.StringAttribute{Computed: true},
			"name":       dschema.StringAttribute{Required: true},
			"code":       dschema.StringAttribute{Computed: true},
			"image_type": dschema.StringAttribute{Computed: true},
			"os":         dschema.StringAttribute{Computed: true},
			"is_public":  dschema.BoolAttribute{Computed: true},
		},
	}
}

func (d *virtualImageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *virtualImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data virtualImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	image, err := d.client.GetVirtualImageByName(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Virtual Image Failed", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(image.ID, 10))
	data.Code = optionalString(image.Code)
	data.ImageType = optionalString(image.ImageType)
	data.OS = optionalString(nestedString(image.OsType, "name"))
	data.IsPublic = maybeBool(image.IsPublic)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// nestedString extracts a string field from a nested JSON object decoded into a
// map[string]any.
func nestedString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
