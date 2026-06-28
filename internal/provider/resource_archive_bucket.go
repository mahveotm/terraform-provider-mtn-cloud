package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

var _ resource.Resource = &archiveBucketResource{}
var _ resource.ResourceWithConfigure = &archiveBucketResource{}
var _ resource.ResourceWithImportState = &archiveBucketResource{}

type archiveBucketResource struct {
	client *client.Client
}

type archiveBucketResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	StorageProvider types.String `tfsdk:"storage_provider"`
	Description     types.String `tfsdk:"description"`
	Visibility      types.String `tfsdk:"visibility"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	AccountID       types.Int64  `tfsdk:"account_id"`
	Code            types.String `tfsdk:"code"`
	FileCount       types.Int64  `tfsdk:"file_count"`
	RawSize         types.Int64  `tfsdk:"raw_size"`
}

func NewArchiveBucketResource() resource.Resource {
	return &archiveBucketResource{}
}

func (r *archiveBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_archive_bucket"
}

func (r *archiveBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceString := []planmodifier.String{stringplanmodifier.RequiresReplace()}

	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud archive bucket backed by a storage provider.",
		Attributes: map[string]rschema.Attribute{
			"id":               rschema.StringAttribute{Computed: true, Description: "Numeric identifier of the archive bucket."},
			"name":             rschema.StringAttribute{Required: true, Description: "Globally unique archive bucket name."},
			"storage_provider": rschema.StringAttribute{Required: true, PlanModifiers: replaceString, Description: "Storage bucket name that backs this archive bucket. Changing it forces a new archive bucket."},
			"description":      rschema.StringAttribute{Optional: true, Description: "Human-readable description of the archive bucket."},
			"visibility": rschema.StringAttribute{
				Optional: true, Computed: true,
				Description: "Archive bucket visibility: `private` or `public`.",
				Validators:  []validator.String{stringvalidator.OneOf("private", "public")},
			},
			"is_public": rschema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the archive bucket is publicly accessible."},
			"account_id": rschema.Int64Attribute{
				Optional:    true,
				Description: "ID of the account that owns the archive bucket.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"code":       rschema.StringAttribute{Computed: true, Description: "Code of the archive bucket."},
			"file_count": rschema.Int64Attribute{Computed: true, Description: "Number of files stored in the archive bucket."},
			"raw_size":   rschema.Int64Attribute{Computed: true, Description: "Total raw size of the archive bucket's contents, in bytes."},
		},
	}
}

func (r *archiveBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := configuredClient(req.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client.")
		return
	}
	r.client = apiClient
}

func (r *archiveBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan archiveBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	provider, err := r.client.GetStorageBucketByName(ctx, plan.StorageProvider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Resolve MTN Cloud Storage Provider Failed", err.Error())
		return
	}
	bucket, err := r.client.CreateArchiveBucket(ctx, archiveBucketInput(plan, provider.ID))
	if err != nil {
		resp.Diagnostics.AddError("Create MTN Cloud Archive Bucket Failed", err.Error())
		return
	}
	setArchiveBucketState(&plan, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *archiveBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state archiveBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Archive Bucket ID", err.Error())
		return
	}
	bucket, err := r.client.GetArchiveBucket(ctx, id)
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Archive Bucket Failed", err.Error())
		return
	}
	setArchiveBucketState(&state, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *archiveBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan archiveBucketResourceModel
	var state archiveBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Archive Bucket ID", err.Error())
		return
	}
	bucket, err := r.client.UpdateArchiveBucket(ctx, id, archiveBucketInput(plan, 0))
	if err != nil {
		resp.Diagnostics.AddError("Update MTN Cloud Archive Bucket Failed", err.Error())
		return
	}
	setArchiveBucketState(&plan, bucket)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *archiveBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state archiveBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Archive Bucket ID", err.Error())
		return
	}
	if err := r.client.DeleteArchiveBucket(ctx, id); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete MTN Cloud Archive Bucket Failed", err.Error())
	}
}

func (r *archiveBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

func archiveBucketInput(plan archiveBucketResourceModel, storageProviderID int64) client.ArchiveBucketInput {
	return client.ArchiveBucketInput{
		Name:              plan.Name.ValueString(),
		StorageProviderID: storageProviderID,
		Description:       plan.Description.ValueString(),
		Visibility:        plan.Visibility.ValueString(),
		IsPublic:          boolPtr(plan.IsPublic),
		AccountID:         int64Ptr(plan.AccountID),
	}
}

func setArchiveBucketState(data *archiveBucketResourceModel, bucket *client.ArchiveBucket) {
	data.ID = types.StringValue(strconv.FormatInt(bucket.ID, 10))
	data.Name = types.StringValue(bucket.Name)
	data.Description = optionalString(bucket.Description)
	data.Visibility = optionalString(bucket.Visibility)
	data.IsPublic = maybeBool(bucket.IsPublic)
	data.Code = optionalString(bucket.Code)
	data.FileCount = maybeInt64(bucket.FileCount)
	data.RawSize = maybeInt64(bucket.RawSize)
}
