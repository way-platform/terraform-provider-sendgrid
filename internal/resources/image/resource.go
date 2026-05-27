package image

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
)

var (
	_ resource.Resource                = (*imageResource)(nil)
	_ resource.ResourceWithImportState = (*imageResource)(nil)
)

type imageResource struct {
	client *client.Client
}

type imageResourceModel struct {
	ID       types.String `tfsdk:"id"`
	FilePath types.String `tfsdk:"file_path"`
	FileHash types.String `tfsdk:"file_sha256"`
	Name     types.String `tfsdk:"name"`
	URL      types.String `tfsdk:"url"`
	Width    types.Int64  `tfsdk:"width"`
	Height   types.Int64  `tfsdk:"height"`
}

// NewResource returns a new image resource constructor.
func NewResource() resource.Resource {
	return &imageResource{}
}

func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *imageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an image uploaded to SendGrid's CDN for use in email templates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The image ID.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"file_path": schema.StringAttribute{
				MarkdownDescription: "Path to the image file to upload.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"file_sha256": schema.StringAttribute{
				MarkdownDescription: "SHA256 hash of the file content. Use `filesha256(...)` to compute. Changes force replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The filename as stored in SendGrid.",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "The CDN URL of the uploaded image.",
				Computed:            true,
			},
			"width": schema.Int64Attribute{
				MarkdownDescription: "Image width in pixels.",
				Computed:            true,
			},
			"height": schema.Int64Attribute{
				MarkdownDescription: "Image height in pixels.",
				Computed:            true,
			},
		},
	}
}

func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	img, err := r.client.UploadImage(ctx, plan.FilePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to upload image", err.Error())
		return
	}

	plan.ID = types.StringValue(img.ID)
	plan.Name = types.StringValue(img.Name)
	plan.URL = types.StringValue(img.URL)
	plan.Width = types.Int64Value(int64(img.Width))
	plan.Height = types.Int64Value(int64(img.Height))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	img, err := r.client.GetImage(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to read image", err.Error())
		return
	}

	state.Name = types.StringValue(img.Name)
	state.URL = types.StringValue(img.URL)
	state.Width = types.Int64Value(int64(img.Width))
	state.Height = types.Int64Value(int64(img.Height))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("update not supported", "image changes force replacement")
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteImage(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("failed to delete image", err.Error())
	}
}

func (r *imageResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	img, err := r.client.GetImage(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import image", err.Error())
		return
	}

	state := imageResourceModel{
		ID:       types.StringValue(img.ID),
		FilePath: types.StringUnknown(),
		FileHash: types.StringUnknown(),
		Name:     types.StringValue(img.Name),
		URL:      types.StringValue(img.URL),
		Width:    types.Int64Value(int64(img.Width)),
		Height:   types.Int64Value(int64(img.Height)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
