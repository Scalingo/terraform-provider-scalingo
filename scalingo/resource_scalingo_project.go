package scalingo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v11"
)

func resourceScalingoProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		DeleteContext: resourceProjectDelete,
		UpdateContext: resourceProjectUpdate,
		CustomizeDiff: resourceProjectCustomizeDiff,

		Description: "Resource representing a project",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "Name of the project",
			},
			"default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not the project is a default project",
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				// Computed keeps the stored slug when the argument is removed from the configuration.
				Computed:         true,
				ValidateDiagFunc: validateProjectOrganization,
				Description: "Slug of the organization in which the project is created. Omit it to create a personal project. " +
					"The slug is only used at creation and is kept in the state as configured: it is not read back from the API. " +
					"Removing the argument has no effect, while adding or changing it on an existing project is refused.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the project",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(scalingo.API)

	params := scalingo.ProjectAddParams{
		Name:    d.Get("name").(string),
		Default: d.Get("default").(bool),
	}

	organizationSlug, _ := d.Get("organization").(string)
	if organizationSlug != "" {
		organizationID, err := projectOrganizationID(ctx, client, organizationSlug)
		if err != nil {
			return diag.Errorf("create project: %v", err)
		}
		params.OrganizationID = organizationID
	}

	project, err := client.ProjectAdd(ctx, params)
	if err != nil {
		return diag.Errorf("create project: %v", err)
	}

	d.SetId(project.ID)
	err = SetAll(d, map[string]interface{}{
		"name":         project.Name,
		"default":      project.Default,
		"organization": organizationSlug,
	})
	if err != nil {
		return diag.Errorf("store project information: %v", err)
	}

	return nil
}

func projectOrganizationID(ctx context.Context, client scalingo.OrganizationsService, organizationSlug string) (string, error) {
	organizations, err := client.OrganizationsList(ctx)
	if err != nil {
		return "", fmt.Errorf("list organizations: %w", err)
	}

	for _, organization := range organizations {
		if organization.Slug == organizationSlug {
			return organization.ID, nil
		}
	}

	return "", fmt.Errorf("organization %q not found", organizationSlug)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(scalingo.API)

	project, err := client.ProjectGet(ctx, d.Id())
	if err != nil {
		return diag.Errorf("get project: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"name":    project.Name,
		"default": project.Default,
	})
	if err != nil {
		return diag.Errorf("store project information: %v", err)
	}

	return nil
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(scalingo.API)

	err := client.ProjectDelete(ctx, d.Id())
	if err != nil {
		return diag.Errorf("remove project: %v", err)
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(scalingo.API)

	// The plan check is skipped when the organization is unknown at plan time.
	if d.HasChange("organization") {
		// Keep the previous state: without it, the SDK stores the refused organization.
		d.Partial(true)
		oldOrganization, newOrganization := d.GetChange("organization")
		return diag.FromErr(projectOrganizationChangeError(oldOrganization.(string), newOrganization.(string)))
	}

	var newProjectNamePtr *string
	// Only set a new project name if a new name was provided
	if newProjectName := d.Get("name").(string); newProjectName != "" {
		newProjectNamePtr = &newProjectName
	}

	var defaultPtr *bool
	// Only allow "default" to be true, otherwise do nothing
	def := d.Get("default").(bool)
	if def {
		defaultPtr = &def
	}

	project, err := client.ProjectUpdate(ctx, d.Id(), scalingo.ProjectUpdateParams{
		Name:    newProjectNamePtr,
		Default: defaultPtr,
	})
	if err != nil {
		return diag.Errorf("update project: %v", err)
	}

	d.SetId(project.ID)
	err = SetAll(d, map[string]interface{}{
		"name":    project.Name,
		"default": project.Default,
	})
	if err != nil {
		return diag.Errorf("store project information: %v", err)
	}

	return nil
}

func resourceProjectCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if d.Id() == "" || !d.NewValueKnown("organization") || !d.HasChange("organization") {
		return nil
	}

	oldOrganization, newOrganization := d.GetChange("organization")
	return projectOrganizationChangeError(oldOrganization.(string), newOrganization.(string))
}

func projectOrganizationChangeError(oldOrganization, newOrganization string) error {
	return fmt.Errorf(
		"organization can only be set when the project is created (stored: %q, configured: %q): restore the stored value or remove the argument",
		oldOrganization, newOrganization,
	)
}

func validateProjectOrganization(value interface{}, path cty.Path) diag.Diagnostics {
	organization, _ := value.(string)
	if strings.TrimSpace(organization) == "" {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       "organization must not be empty",
			Detail:        "Provide an organization slug, or omit the argument to create a personal project.",
			AttributePath: path,
		}}
	}
	return nil
}
