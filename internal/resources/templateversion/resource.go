package templateversion

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

var (
	_ resource.Resource                = (*templateVersionResource)(nil)
	_ resource.ResourceWithImportState = (*templateVersionResource)(nil)
)

type templateVersionResource struct {
	client *client.Client
}

type templateVersionResourceModel struct {
	ID         types.String `tfsdk:"id"`
	TemplateID types.String `tfsdk:"template_id"`
	Name       types.String `tfsdk:"name"`
	Subject    types.String `tfsdk:"subject"`
	HTML       types.String `tfsdk:"html_content"`
	Active     types.Int64  `tfsdk:"active"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// NewResource returns a new template version resource constructor.
func NewResource() resource.Resource {
	return &templateVersionResource{}
}

func (r *templateVersionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_template_version"
}

func (r *templateVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a version of a SendGrid dynamic template. Versions are immutable — changing content forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The version ID.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the parent template.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of this version.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subject": schema.StringAttribute{
				MarkdownDescription: "The subject line (supports Handlebars).",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"html_content": schema.StringAttribute{
				MarkdownDescription: "The HTML body of the email (supports Handlebars).",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"active": schema.Int64Attribute{
				MarkdownDescription: "Whether this version is active (1) or inactive (0).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last time this version was updated.",
				Computed:            true,
			},
		},
	}
}

func (r *templateVersionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", "expected *client.Client")
		return
	}
	r.client = c
}

func (r *templateVersionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan templateVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	active := templates.Active(plan.Active.ValueInt64())
	editor := templates.EditorCode
	tv, err := r.client.CreateTemplateVersion(
		ctx,
		plan.TemplateID.ValueString(),
		templates.CreateTemplateVersionJSONRequestBody{
			Name:        plan.Name.ValueString(),
			Subject:     plan.Subject.ValueString(),
			HTMLContent: client.Ptr(plan.HTML.ValueString()),
			Active:      &active,
			Editor:      &editor,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create template version", err.Error())
		return
	}

	if tv.ID == nil {
		resp.Diagnostics.AddError("invalid response", "template version ID is missing from API response")
		return
	}
	plan.ID = types.StringValue(*tv.ID)
	if tv.UpdatedAt != nil {
		plan.UpdatedAt = types.StringValue(*tv.UpdatedAt)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tv, err := r.client.GetTemplateVersion(ctx, state.TemplateID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read template version", err.Error())
		return
	}

	state.Name = types.StringValue(tv.Name)
	state.Subject = types.StringValue(tv.Subject)
	if tv.HTMLContent != nil {
		state.HTML = types.StringValue(*tv.HTMLContent)
	}
	if tv.Active != nil {
		state.Active = types.Int64Value(int64(*tv.Active))
	}
	if tv.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(*tv.UpdatedAt)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateVersionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("update not supported", "template versions are immutable; changes force replacement")
}

func (r *templateVersionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state templateVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTemplateVersion(ctx, state.TemplateID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("failed to delete template version", err.Error())
	}
}

func (r *templateVersionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import ID", "expected format: <template_id>/<version_id>")
		return
	}
	templateID, versionID := parts[0], parts[1]

	tv, err := r.client.GetTemplateVersion(ctx, templateID, versionID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import template version", err.Error())
		return
	}

	if tv.ID == nil {
		resp.Diagnostics.AddError("invalid response", "template version ID is missing from API response")
		return
	}
	state := templateVersionResourceModel{
		ID:         types.StringValue(*tv.ID),
		TemplateID: types.StringValue(templateID),
		Name:       types.StringValue(tv.Name),
		Subject:    types.StringValue(tv.Subject),
	}
	if tv.HTMLContent != nil {
		state.HTML = types.StringValue(*tv.HTMLContent)
	}
	if tv.Active != nil {
		state.Active = types.Int64Value(int64(*tv.Active))
	}
	if tv.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(*tv.UpdatedAt)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
