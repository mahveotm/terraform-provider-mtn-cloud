package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

var _ resource.Resource = &wikiPageResource{}
var _ resource.ResourceWithConfigure = &wikiPageResource{}
var _ resource.ResourceWithImportState = &wikiPageResource{}
var _ basetypes.StringTypable = wikiContentType{}
var _ basetypes.StringValuableWithSemanticEquals = wikiContentValue{}

type wikiContentType struct {
	basetypes.StringType
}

func (t wikiContentType) Equal(o attr.Type) bool {
	_, ok := o.(wikiContentType)
	return ok
}

func (t wikiContentType) String() string {
	return "wikiContentType"
}

func (t wikiContentType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return wikiContentValue{StringValue: in}, nil
}

func (t wikiContentType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	value, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := value.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", value)
	}
	return wikiContentValue{StringValue: stringValue}, nil
}

func (t wikiContentType) ValueType(_ context.Context) attr.Value {
	return wikiContentValue{}
}

type wikiContentValue struct {
	basetypes.StringValue
}

func newWikiContentValue(value string) wikiContentValue {
	return wikiContentValue{StringValue: types.StringValue(value)}
}

func (v wikiContentValue) Equal(o attr.Value) bool {
	other, ok := o.(wikiContentValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v wikiContentValue) Type(_ context.Context) attr.Type {
	return wikiContentType{}
}

func (v wikiContentValue) StringSemanticEquals(ctx context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	otherValue, otherDiags := other.ToStringValue(ctx)
	diags.Append(otherDiags...)
	if diags.HasError() {
		return false, diags
	}
	if v.IsNull() || v.IsUnknown() || otherValue.IsNull() || otherValue.IsUnknown() {
		return false, diags
	}
	return trimSingleTrailingNewline(v.ValueString()) == trimSingleTrailingNewline(otherValue.ValueString()), diags
}

func trimSingleTrailingNewline(value string) string {
	return strings.TrimSuffix(value, "\n")
}

type wikiPageResource struct {
	resourceBase
}

type wikiPageResourceModel struct {
	ID       types.String     `tfsdk:"id"`
	Name     types.String     `tfsdk:"name"`
	Category types.String     `tfsdk:"category"`
	Content  wikiContentValue `tfsdk:"content"`
}

func NewWikiPageResource() resource.Resource {
	return &wikiPageResource{}
}

func (r *wikiPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wiki_page"
}

func (r *wikiPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud wiki page.",
		Attributes: map[string]rschema.Attribute{
			"id":       computedIDAttribute("Numeric identifier of the wiki page."),
			"name":     rschema.StringAttribute{Required: true, Description: "Name (title) of the wiki page."},
			"category": rschema.StringAttribute{Optional: true, Computed: true, Description: "Category the wiki page belongs to."},
			"content": rschema.StringAttribute{
				Optional:    true,
				CustomType:  wikiContentType{},
				Description: "Markdown content of the wiki page.",
			},
		},
	}
}

func (r *wikiPageResource) input(plan wikiPageResourceModel) client.WikiPageInput {
	return client.WikiPageInput{
		Name:     plan.Name.ValueString(),
		Category: plan.Category.ValueString(),
		Content:  plan.Content.ValueString(),
	}
}

func (r *wikiPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wikiPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := r.client.CreateWikiPage(ctx, r.input(plan))
	if err != nil {
		opError(&resp.Diagnostics, "Create", "Wiki Page", err)
		return
	}
	setWikiPageState(&plan, page)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wikiPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wikiPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "Wiki Page", &resp.Diagnostics)
	if !ok {
		return
	}
	page, err := r.client.GetWikiPage(ctx, id)
	if handleReadError(ctx, err, "Wiki Page", &resp.State, &resp.Diagnostics) {
		return
	}
	setWikiPageState(&state, page)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wikiPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wikiPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(plan.ID, "Wiki Page", &resp.Diagnostics)
	if !ok {
		return
	}
	page, err := r.client.UpdateWikiPage(ctx, id, r.input(plan))
	if err != nil {
		opError(&resp.Diagnostics, "Update", "Wiki Page", err)
		return
	}
	setWikiPageState(&plan, page)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wikiPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state wikiPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, ok := parseID(state.ID, "Wiki Page", &resp.Diagnostics)
	if !ok {
		return
	}
	if err := r.client.DeleteWikiPage(ctx, id); err != nil && !client.IsNotFound(err) {
		opError(&resp.Diagnostics, "Delete", "Wiki Page", err)
	}
}

func (r *wikiPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

func setWikiPageState(data *wikiPageResourceModel, page *client.WikiPage) {
	data.ID = types.StringValue(strconv.FormatInt(page.ID, 10))
	data.Name = types.StringValue(page.Name)
	data.Category = mergeAPIString(data.Category, page.Category)
	// content stays Optional (not Computed); only overwrite when the API
	// returns a value so an unset content does not flip to a populated string.
	if page.Content != "" {
		data.Content = newWikiContentValue(page.Content)
	}
}
