package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vezor/terraform-provider-vezor/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SecretDataSource{}

// SecretDataSource defines the data source implementation
type SecretDataSource struct {
	client *client.Client
}

// SecretDataSourceModel describes the data source data model
type SecretDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
	Tags        types.Map    `tfsdk:"tags"`
	Version     types.Int64  `tfsdk:"version"`
}

// NewSecretDataSource creates a new secret data source
func NewSecretDataSource() datasource.DataSource {
	return &SecretDataSource{}
}

// Metadata returns the data source type name
func (d *SecretDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

// Schema defines the schema for the data source
func (d *SecretDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single secret from Vezor by name and tags.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the secret.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name (key) of the secret to fetch.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The decrypted value of the secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the secret.",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Tags to filter the secret. At minimum, 'env' and 'app' are typically required.",
				Required:    true,
				ElementType: types.StringType,
			},
			"version": schema.Int64Attribute{
				Description: "The version number of the secret.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *SecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data
func (d *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecretDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert tags from Terraform types to Go map
	tags := make(map[string]string)
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the secret from the API
	secret, err := d.client.FindSecret(data.Name.ValueString(), tags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Secret",
			fmt.Sprintf("Unable to read secret '%s': %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(secret.ID)
	data.Name = types.StringValue(secret.KeyName)
	data.Value = types.StringValue(secret.Value)
	data.Description = types.StringValue(secret.Description)
	data.Version = types.Int64Value(int64(secret.Version))

	// Convert tags back to Terraform types
	tagsMap, diags := types.MapValueFrom(ctx, types.StringType, secret.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tagsMap

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
