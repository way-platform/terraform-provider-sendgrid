package template

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/way-platform/terraform-provider-sendgrid/api/templates"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

var (
	_ resource.Resource                = (*templateResource)(nil)
	_ resource.ResourceWithImportState = (*templateResource)(nil)
)

type templateResource struct {
	client *client.Client
}

type templateResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Generation types.String `tfsdk:"generation"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// NewResource returns a new template resource constructor.
func NewResource() resource.Resource {
	return &templateResource{}
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SendGrid dynamic template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The template ID (e.g. `d-xxxx`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the template.",
				Required:            true,
			},
			"generation": schema.StringAttribute{
				MarkdownDescription: "The generation of the template. Must be `dynamic`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dynamic"),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last time the template was updated.",
				Computed:            true,
			},
		},
	}
}

func (r *templateResource) Configure(
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

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gen := templates.Generation(plan.Generation.ValueString())
	t, err := r.client.CreateTemplate(ctx, templates.CreateTemplateJSONRequestBody{
		Name:       plan.Name.ValueString(),
		Generation: &gen,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create template", err.Error())
		return
	}

	plan.ID = types.StringValue(t.ID)
	plan.UpdatedAt = types.StringValue(t.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.client.GetTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read template", err.Error())
		return
	}

	state.Name = types.StringValue(t.Name)
	state.Generation = types.StringValue(string(t.Generation))
	state.UpdatedAt = types.StringValue(t.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.client.UpdateTemplate(ctx, state.ID.ValueString(), templates.UpdateTemplateJSONRequestBody{
		Name: client.Ptr(plan.Name.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update template", err.Error())
		return
	}

	plan.ID = state.ID
	plan.UpdatedAt = types.StringValue(t.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTemplate(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("failed to delete template", err.Error())
	}
}

func (r *templateResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	t, err := r.client.GetTemplate(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import template", err.Error())
		return
	}

	state := templateResourceModel{
		ID:         types.StringValue(t.ID),
		Name:       types.StringValue(t.Name),
		Generation: types.StringValue(string(t.Generation)),
		UpdatedAt:  types.StringValue(t.UpdatedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
