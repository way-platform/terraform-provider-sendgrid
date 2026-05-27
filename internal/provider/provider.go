package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/way-platform/terraform-provider-sendgrid/internal/client"
	"github.com/way-platform/terraform-provider-sendgrid/internal/resources/image"
	"github.com/way-platform/terraform-provider-sendgrid/internal/resources/template"
	"github.com/way-platform/terraform-provider-sendgrid/internal/resources/templateversion"
)

var _ provider.Provider = (*sendGridProvider)(nil)

type sendGridProvider struct {
	version string
}

type sendGridProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &sendGridProvider{version: version}
	}
}

func (p *sendGridProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sendgrid"
	resp.Version = p.version
}

func (p *sendGridProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The SendGrid provider manages SendGrid email templates and images.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "SendGrid API key. Can also be set via the `SENDGRID_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *sendGridProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var config sendGridProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("SENDGRID_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"missing API key",
			"Set the api_key provider attribute or the SENDGRID_API_KEY environment variable.",
		)
		return
	}

	var opts []client.Option
	if baseURL := os.Getenv("SENDGRID_API_URL"); baseURL != "" {
		opts = append(opts, client.WithBaseURL(baseURL))
	}

	c := client.New(apiKey, opts...)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *sendGridProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		template.NewResource,
		templateversion.NewResource,
		image.NewResource,
	}
}

func (p *sendGridProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
