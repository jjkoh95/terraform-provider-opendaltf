package provider

import (
	"context"
	"fmt"
	"strings"

	opendal "github.com/apache/opendal/bindings/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &opendalBucketResource{}
	_ resource.ResourceWithConfigure   = &opendalBucketResource{}
	_ resource.ResourceWithImportState = &opendalBucketResource{}
)

// opendalBucketResourceModel maps the resource schema data.
type opendalBucketResourceModel struct {
	Name types.String `tfsdk:"name"`
}

// opendalBucketResource defines the resource implementation.
type opendalBucketResource struct {
	Operator *opendal.Operator
}

// NewOpendalBucketResource is a helper function to simplify the provider implementation.
func NewOpendalBucketResource() resource.Resource {
	return &opendalBucketResource{}
}

// Metadata returns the resource type name.
func (r *opendalBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

// Schema defines the schema for the resource.
func (r *opendalBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a bucket (directory) using Apache OpenDAL.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the bucket (directory).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(), // Changing the name requires recreating
				},
			},
			// Add other potential bucket configurations here if needed later
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *opendalBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*opendalTFProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *opendalTFProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.Operator = providerData.Operator
}

// ensureDirPath ensures the path ends with a slash, common for directories in OpenDAL.
func ensureDirPath(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

// Create creates the resource and sets the initial Terraform state.
func (r *opendalBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var plan opendalBucketResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucketName := plan.Name.ValueString()
	dirPath := ensureDirPath(bucketName)

	tflog.Debug(ctx, fmt.Sprintf("Creating directory: %s", dirPath))
	err := r.Operator.CreateDir(dirPath)
	if err != nil {
		// Basic error handling, could add checks for existing dirs later
		resp.Diagnostics.AddError(
			"Error Creating OpenDAL Bucket",
			fmt.Sprintf("Could not create bucket '%s', unexpected error: %s", bucketName, err.Error()),
		)
		return
	}

	// Set state to reflect creation
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *opendalBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// For now, we'll just assume if the resource exists in state, it exists upstream.
	// A real implementation would use Stat or List to verify.
	tflog.Debug(ctx, "Read operation called, assuming resource exists if in state (basic implementation)")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *opendalBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// "name" requires replacement, so Update shouldn't be called for it.
	tflog.Debug(ctx, "Update operation called, but no updatable attributes are defined.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *opendalBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// For now, we'll assume deletion is successful.
	// A real implementation would call Operator.Delete(dirPath) or Operator.RemoveAll(dirPath).
	tflog.Debug(ctx, "Delete operation called, removing resource from state (basic implementation)")
}

// ImportState imports the resource into Terraform state.
func (r *opendalBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID provided (which should be the bucket name) to set the name attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
