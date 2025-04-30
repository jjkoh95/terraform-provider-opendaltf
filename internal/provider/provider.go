package provider

import (
	"context"

	"github.com/apache/opendal-go-services/fs"
	"github.com/apache/opendal-go-services/gcs"
	opendal "github.com/apache/opendal/bindings/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &opendalTFProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &opendalTFProvider{
			version: version,
		}
	}
}

// opendalTFProvider is the provider implementation.
type opendalTFProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	Operator *opendal.Operator
}

// Metadata returns the provider type name.
func (p *opendalTFProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "opendaltf"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *opendalTFProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_type": schema.StringAttribute{
				Description: "The type of the storage service (e.g., 'fs', 's3', 'azblob', 'gcs').",
				Required:    true,
			},
			"config": schema.MapAttribute{
				Description: "Configuration map for the OpenDAL service (e.g., root, bucket, endpoint, region, credentials).",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

// Configure prepares a OpenDAL TF API client for data sources and resources.
func (p *opendalTFProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config struct {
		ServiceType types.String `tfsdk:"service_type"`
		Config      types.Map    `tfsdk:"config"`
	}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceType := config.ServiceType.ValueString()
	configMap := make(map[string]string)
	if !config.Config.IsNull() {
		diags := config.Config.ElementsAs(ctx, &configMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Map service type string to OpenDAL Scheme (add more cases as needed)
	var scheme opendal.Scheme
	switch serviceType {
	case "gcs":
		scheme = gcs.Scheme
	case "fs":
		scheme = fs.Scheme
	// Add cases for other services like "s3", "fs", "azblob", etc.
	default:
		resp.Diagnostics.AddError(
			"Unsupported Service Type",
			"The provided service_type '"+serviceType+"' is not supported by this provider.",
		)
		return
	}

	op, err := opendal.NewOperator(scheme, configMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create OpenDAL Operator",
			"Could not create OpenDAL operator for service '"+serviceType+"': "+err.Error(),
		)
		return
	}

	p.Operator = op

	// Make the operator available to resources and data sources
	resp.DataSourceData = p
	resp.ResourceData = p
}

// DataSources defines the data sources implemented in the provider.
func (p *opendalTFProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *opendalTFProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOpendalBucketResource,
	}
}
