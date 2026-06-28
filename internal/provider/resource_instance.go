package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mahveotm/terraform-provider-mtn-cloud/internal/client"
)

var _ resource.Resource = &instanceResource{}
var _ resource.ResourceWithConfigure = &instanceResource{}
var _ resource.ResourceWithImportState = &instanceResource{}

type instanceResource struct {
	client   *client.Client
	defaults *mtnCloudProviderData
}

type instanceResourceModel struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Group               types.String   `tfsdk:"group"`
	Type                types.String   `tfsdk:"type"`
	Plan                types.String   `tfsdk:"plan"`
	ResourcePool        types.String   `tfsdk:"resource_pool"`
	Description         types.String   `tfsdk:"description"`
	Environment         types.String   `tfsdk:"environment"`
	Labels              types.List     `tfsdk:"labels"`
	LabelsAll           types.List     `tfsdk:"labels_all"`
	Tags                types.Map      `tfsdk:"tags"`
	TagsAll             types.Map      `tfsdk:"tags_all"`
	AvailabilityZone    types.String   `tfsdk:"availability_zone"`
	SecurityGroup       types.String   `tfsdk:"security_group"`
	SecurityGroups      types.List     `tfsdk:"security_groups"`
	OSExternalNetworkID types.String   `tfsdk:"os_external_network_id"`
	CreateUser          types.Bool     `tfsdk:"create_user"`
	WorkflowID          types.Int64    `tfsdk:"workflow_id"`
	ShutdownDays        types.Int64    `tfsdk:"shutdown_days"`
	ExpireDays          types.Int64    `tfsdk:"expire_days"`
	CreateBackup        types.Bool     `tfsdk:"create_backup"`
	WaitForReady        types.Bool     `tfsdk:"wait_for_ready"`
	Status              types.String   `tfsdk:"status"`
	PrimaryIP           types.String   `tfsdk:"primary_ip"`
	ExternalIP          types.String   `tfsdk:"external_ip"`
	CloudID             types.Int64    `tfsdk:"cloud_id"`
	GroupID             types.Int64    `tfsdk:"group_id"`
	LayoutID            types.Int64    `tfsdk:"layout_id"`
	PlanID              types.Int64    `tfsdk:"plan_id"`
	ResourcePoolID      types.String   `tfsdk:"resource_pool_id"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

func NewInstanceResource() resource.Resource {
	return &instanceResource{}
}

func (r *instanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *instanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceString := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceBool := []planmodifier.Bool{boolplanmodifier.RequiresReplace()}
	replaceInt64 := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	// Inherited Optional+Computed fields: force replacement when changed, but keep
	// the resolved value across updates when left unset.
	inheritString := []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}

	resp.Schema = rschema.Schema{
		Description: "Manages an MTN Cloud compute instance using human-friendly provisioning inputs.",
		Attributes: map[string]rschema.Attribute{
			"id":                     rschema.StringAttribute{Computed: true},
			"name":                   rschema.StringAttribute{Required: true, PlanModifiers: replaceString},
			"group":                  rschema.StringAttribute{Optional: true, Computed: true, PlanModifiers: inheritString, Description: "Group/site name. Defaults to the provider's `group`."},
			"type":                   rschema.StringAttribute{Required: true, PlanModifiers: replaceString},
			"plan":                   rschema.StringAttribute{Required: true},
			"resource_pool":          rschema.StringAttribute{Optional: true, Computed: true, PlanModifiers: inheritString, Description: "Resource pool name/code. Defaults to the provider's `resource_pool`; if neither is set and the group has exactly one pool, that pool is used."},
			"description":            rschema.StringAttribute{Optional: true},
			"environment":            rschema.StringAttribute{Optional: true},
			"availability_zone":      rschema.StringAttribute{Optional: true, Computed: true, PlanModifiers: inheritString, Description: "Availability zone. Defaults to the provider's `availability_zone`."},
			"security_group":         rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("default")},
			"os_external_network_id": rschema.StringAttribute{Optional: true, PlanModifiers: replaceString},
			"create_user":            rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), PlanModifiers: replaceBool},
			"workflow_id":            rschema.Int64Attribute{Optional: true, PlanModifiers: replaceInt64, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"shutdown_days":          rschema.Int64Attribute{Optional: true, PlanModifiers: replaceInt64, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"expire_days":            rschema.Int64Attribute{Optional: true, PlanModifiers: replaceInt64, Validators: []validator.Int64{int64validator.AtLeast(1)}},
			"create_backup":          rschema.BoolAttribute{Optional: true, PlanModifiers: replaceBool},
			"wait_for_ready":         rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"labels": rschema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"labels_all": rschema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Effective labels: the provider's default_labels merged (union) with `labels`.",
			},
			"tags": rschema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"tags_all": rschema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Effective tags: the provider's default_tags merged with `tags` (resource values win).",
			},
			"security_groups": rschema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"status":           rschema.StringAttribute{Computed: true},
			"primary_ip":       rschema.StringAttribute{Computed: true},
			"external_ip":      rschema.StringAttribute{Computed: true},
			"cloud_id":         rschema.Int64Attribute{Computed: true},
			"group_id":         rschema.Int64Attribute{Computed: true},
			"layout_id":        rschema.Int64Attribute{Computed: true},
			"plan_id":          rschema.Int64Attribute{Computed: true},
			"resource_pool_id": rschema.StringAttribute{Computed: true},
		},
		Blocks: map[string]rschema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := configuredProvider(req.ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *mtnCloudProviderData.")
		return
	}
	r.client = data.Client
	r.defaults = data
}

func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout, diags := plan.Timeouts.Create(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	input, resolved, err := r.createInput(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Resolve MTN Cloud Instance Inputs Failed", err.Error())
		return
	}
	// Record the resolved values for the inherited Optional+Computed inputs.
	plan.Group = types.StringValue(resolved.GroupName)
	plan.ResourcePool = types.StringValue(resolved.ResourcePoolName)
	plan.AvailabilityZone = optionalString(resolved.AvailabilityZone)
	resp.Diagnostics.Append(r.setComputedTags(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.client.ProvisionInstance(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create MTN Cloud Instance Failed", err.Error())
		return
	}
	if plan.WaitForReady.ValueBool() {
		instance, err = r.client.WaitForInstanceStatus(ctx, instance.ID, "running", 5*time.Second)
		if err != nil {
			resp.Diagnostics.AddError("Wait for MTN Cloud Instance Failed", err.Error())
			return
		}
	}
	setInstanceState(&plan, instance, resolved)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	readTimeout, diags := state.Timeouts.Read(ctx, 2*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Instance ID", err.Error())
		return
	}
	instance, err := r.client.GetInstance(ctx, id)
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read MTN Cloud Instance Failed", err.Error())
		return
	}
	setInstanceObservedState(&state, instance)
	resp.Diagnostics.Append(r.setComputedTags(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceResourceModel
	var state instanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateTimeout, diags := plan.Timeouts.Update(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Instance ID", err.Error())
		return
	}
	if !plan.Plan.Equal(state.Plan) {
		resolved, err := r.resolveProvisioning(ctx, plan)
		if err != nil {
			resp.Diagnostics.AddError("Resolve MTN Cloud Plan Failed", err.Error())
			return
		}
		if err := r.client.ResizeInstance(ctx, id, resolved.PlanID); err != nil {
			resp.Diagnostics.AddError("Resize MTN Cloud Instance Failed", err.Error())
			return
		}
		plan.PlanID = types.Int64Value(resolved.PlanID)
	}

	labels := mergeLabels(r.defaults.DefaultLabels, stringList(ctx, plan.Labels))
	updated, err := r.client.UpdateInstance(ctx, id, client.UpdateInstanceInput{
		Description: stringPtr(plan.Description),
		Labels:      labels,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update MTN Cloud Instance Failed", err.Error())
		return
	}
	plan.ID = state.ID
	plan.CloudID = state.CloudID
	plan.GroupID = state.GroupID
	plan.LayoutID = state.LayoutID
	plan.ResourcePoolID = state.ResourcePoolID
	setInstanceObservedState(&plan, updated)
	resp.Diagnostics.Append(r.setComputedTags(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, diags := state.Timeouts.Delete(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Instance ID", err.Error())
		return
	}
	if err := r.client.DeleteInstance(ctx, id); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete MTN Cloud Instance Failed", err.Error())
	}
}

func (r *instanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID(), req, resp)
}

type resolvedProvisioning struct {
	GroupName        string
	CloudName        string
	CloudID          int64
	GroupID          int64
	LayoutID         int64
	PlanID           int64
	ResourcePoolID   string // normalized code sent to the API
	ResourcePoolName string // value written back to state
	AvailabilityZone string
}

func (r *instanceResource) createInput(ctx context.Context, plan instanceResourceModel) (client.CreateInstanceInput, resolvedProvisioning, error) {
	resolved, err := r.resolveProvisioning(ctx, plan)
	if err != nil {
		return client.CreateInstanceInput{}, resolvedProvisioning{}, err
	}
	return client.CreateInstanceInput{
		Name:                plan.Name.ValueString(),
		Cloud:               resolved.CloudName,
		Type:                plan.Type.ValueString(),
		GroupID:             resolved.GroupID,
		LayoutID:            resolved.LayoutID,
		PlanID:              resolved.PlanID,
		ResourcePoolID:      resolved.ResourcePoolID,
		Description:         plan.Description.ValueString(),
		Environment:         plan.Environment.ValueString(),
		Labels:              mergeLabels(r.defaults.DefaultLabels, stringList(ctx, plan.Labels)),
		Tags:                mergeTags(r.defaults.DefaultTags, stringMap(ctx, plan.Tags)),
		AvailabilityZone:    resolved.AvailabilityZone,
		SecurityGroup:       plan.SecurityGroup.ValueString(),
		SecurityGroups:      stringList(ctx, plan.SecurityGroups),
		OSExternalNetworkID: plan.OSExternalNetworkID.ValueString(),
		CreateUser:          boolPtr(plan.CreateUser),
		WorkflowID:          int64Ptr(plan.WorkflowID),
		ShutdownDays:        int64Ptr(plan.ShutdownDays),
		ExpireDays:          int64Ptr(plan.ExpireDays),
		CreateBackup:        boolPtr(plan.CreateBackup),
	}, resolved, nil
}

func (r *instanceResource) resolveProvisioning(ctx context.Context, plan instanceResourceModel) (resolvedProvisioning, error) {
	groupName := firstNonEmpty(plan.Group.ValueString(), r.defaults.Group)
	if groupName == "" {
		return resolvedProvisioning{}, fmt.Errorf("`group` is required: set it on the resource or as a provider-level default")
	}
	group, instanceType, err := resolveGroupAndInstanceType(ctx, r.client, groupName, plan.Type.ValueString())
	if err != nil {
		return resolvedProvisioning{}, err
	}
	if len(group.CloudIDs) == 0 {
		return resolvedProvisioning{}, fmt.Errorf("group %q has no cloud IDs", group.Name)
	}
	if instanceType.DefaultLayoutID == nil {
		return resolvedProvisioning{}, fmt.Errorf("instance type %q has no default layout ID", instanceType.Code)
	}
	planResult, err := r.client.GetServicePlan(ctx, plan.Plan.ValueString(), group.CloudIDs[0], *instanceType.DefaultLayoutID, group.ID)
	if err != nil {
		return resolvedProvisioning{}, err
	}

	pool, err := r.resolveResourcePool(ctx, plan, group)
	if err != nil {
		return resolvedProvisioning{}, err
	}

	return resolvedProvisioning{
		GroupName:        groupName,
		CloudName:        group.Name,
		CloudID:          group.CloudIDs[0],
		GroupID:          group.ID,
		LayoutID:         *instanceType.DefaultLayoutID,
		PlanID:           planResult.ID,
		ResourcePoolID:   client.NormalizeResourcePoolID(pool.Code),
		ResourcePoolName: firstNonEmpty(plan.ResourcePool.ValueString(), pool.Name, pool.Code),
		AvailabilityZone: firstNonEmpty(plan.AvailabilityZone.ValueString(), r.defaults.AvailabilityZone),
	}, nil
}

// resolveResourcePool honours an explicit pool (resource or provider default),
// and otherwise auto-selects when the group exposes exactly one pool.
func (r *instanceResource) resolveResourcePool(ctx context.Context, plan instanceResourceModel, group *client.Group) (*client.ResourcePool, error) {
	if name := firstNonEmpty(plan.ResourcePool.ValueString(), r.defaults.ResourcePool); name != "" {
		return r.client.GetResourcePool(ctx, name, group)
	}
	pools, err := r.client.ListResourcePools(ctx, group.CloudIDs[0], group.ID)
	if err != nil {
		return nil, err
	}
	switch len(pools) {
	case 0:
		return nil, fmt.Errorf("group %q has no resource pools; set `resource_pool`", group.Name)
	case 1:
		return &pools[0], nil
	default:
		names := make([]string, 0, len(pools))
		for _, pool := range pools {
			names = append(names, pool.Name)
		}
		return nil, fmt.Errorf(
			"group %q has %d resource pools; set `resource_pool` (or the provider default) to one of: %s",
			group.Name, len(pools), strings.Join(names, ", "),
		)
	}
}

func setInstanceState(data *instanceResourceModel, instance *client.Instance, resolved resolvedProvisioning) {
	data.CloudID = types.Int64Value(resolved.CloudID)
	data.GroupID = types.Int64Value(resolved.GroupID)
	data.LayoutID = types.Int64Value(resolved.LayoutID)
	data.PlanID = types.Int64Value(resolved.PlanID)
	data.ResourcePoolID = types.StringValue(resolved.ResourcePoolID)
	setInstanceObservedState(data, instance)
}

func setInstanceObservedState(data *instanceResourceModel, instance *client.Instance) {
	data.ID = types.StringValue(strconv.FormatInt(instance.ID, 10))
	data.Status = types.StringValue(instance.Status)
	data.PrimaryIP = types.StringValue(firstNonEmpty(instance.IPAddress, instance.ExternalIP))
	data.ExternalIP = types.StringValue(instance.ExternalIP)
}

// setComputedTags fills labels_all / tags_all with the effective values
// (provider defaults merged with the resource's own labels/tags).
func (r *instanceResource) setComputedTags(ctx context.Context, data *instanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	labels := mergeLabels(r.defaults.DefaultLabels, stringList(ctx, data.Labels))
	tags := mergeTags(r.defaults.DefaultTags, stringMap(ctx, data.Tags))

	labelsAll, d := types.ListValueFrom(ctx, types.StringType, labels)
	diags.Append(d...)
	tagsAll, d := types.MapValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)

	data.LabelsAll = labelsAll
	data.TagsAll = tagsAll
	return diags
}

func stringList(ctx context.Context, value types.List) []string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var result []string
	_ = value.ElementsAs(ctx, &result, false)
	return result
}

func stringMap(ctx context.Context, value types.Map) map[string]string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var result map[string]string
	_ = value.ElementsAs(ctx, &result, false)
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
