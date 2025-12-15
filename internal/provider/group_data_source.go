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
var _ datasource.DataSource = &GroupDataSource{}

// GroupDataSource defines the data source implementation
type GroupDataSource struct {
	client *client.Client
}

// GroupDataSourceModel describes the data source data model
type GroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.Map    `tfsdk:"tags"`
	Secrets     types.Map    `tfsdk:"secrets"`
	SecretCount types.Int64  `tfsdk:"secret_count"`
}

// NewGroupDataSource creates a new group data source
func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

// Metadata returns the data source type name
func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the data source
func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all secrets from a Vezor group. Groups are saved tag queries that match multiple secrets.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the group.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the group to fetch secrets from.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the group.",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "The tags that define this group's query.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"secrets": schema.MapAttribute{
				Description: "A map of secret names to their decrypted values. Can be used directly with kubernetes_secret or other resources.",
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"secret_count": schema.Int64Attribute{
				Description: "The number of secrets in this group.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := data.Name.ValueString()

	// Fetch the group metadata
	group, err := d.client.GetGroup(groupName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Group",
			fmt.Sprintf("Unable to read group '%s': %s", groupName, err.Error()),
		)
		return
	}

	// Fetch the secrets for this group
	groupSecrets, err := d.client.PullGroupSecrets(groupName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Pull Group Secrets",
			fmt.Sprintf("Unable to pull secrets for group '%s': %s", groupName, err.Error()),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)
	data.SecretCount = types.Int64Value(int64(groupSecrets.Count))

	// Convert tags to Terraform types
	tagsMap, diags := types.MapValueFrom(ctx, types.StringType, group.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Tags = tagsMap

	// Convert secrets to Terraform types
	secretsMap, diags := types.MapValueFrom(ctx, types.StringType, groupSecrets.Secrets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Secrets = secretsMap

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
