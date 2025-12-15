package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vezor/terraform-provider-vezor/internal/client"
)

// Ensure VezorProvider satisfies various provider interfaces
var _ provider.Provider = &VezorProvider{}

// VezorProvider defines the provider implementation
type VezorProvider struct {
	version string
}

// VezorProviderModel describes the provider data model
type VezorProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIURL types.String `tfsdk:"api_url"`
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VezorProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *VezorProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vezor"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *VezorProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Vezor provider allows you to access secrets stored in Vezor.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "The API key for authenticating with Vezor. Can also be set via the VEZOR_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "The URL of the Vezor API. Defaults to https://api.vezor.io. Can also be set via the VEZOR_API_URL environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a Vezor API client for data sources and resources
func (p *VezorProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config VezorProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get API key from config or environment
	apiKey := os.Getenv("VEZOR_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires an API key. Set it in the provider configuration or via the VEZOR_API_KEY environment variable.",
		)
		return
	}

	// Get API URL from config or environment, default to production
	apiURL := "https://api.vezor.io"
	if envURL := os.Getenv("VEZOR_API_URL"); envURL != "" {
		apiURL = envURL
	}
	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()
	}

	// Create the API client
	vezorClient := client.NewClient(apiURL, apiKey)

	// Make the client available to data sources and resources
	resp.DataSourceData = vezorClient
	resp.ResourceData = vezorClient
}

// Resources defines the resources implemented in the provider
func (p *VezorProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Resources would go here if we implement secret management
	}
}

// DataSources defines the data sources implemented in the provider
func (p *VezorProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSecretDataSource,
		NewGroupDataSource,
	}
}
