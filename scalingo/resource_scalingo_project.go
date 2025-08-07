package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func resourceScalingoProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		DeleteContext: resourceProjectDelete,
		UpdateContext: resourceProjectUpdate,

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
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the project",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectImport,
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	project, err := client.ProjectAdd(ctx, scalingo.ProjectAddParams{
		Name:    d.Get("name").(string),
		Default: d.Get("default").(bool),
	})
	if err != nil {
		return diag.Errorf("create project: %v", err)
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

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

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
	client, _ := meta.(*scalingo.Client)

	err := client.ProjectDelete(ctx, d.Id())
	if err != nil {
		return diag.Errorf("remove project: %v", err)
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

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

func resourceProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, _ := meta.(*scalingo.Client)

	if d.Id() == "" {
		return nil, fmt.Errorf("project ID is empty")
	}

	project, err := client.ProjectGet(ctx, d.Id())
	if err != nil {
		return nil, fmt.Errorf("get project: %v", err)
	}

	d.SetId(project.ID)
	err = SetAll(d, map[string]interface{}{
		"name":    project.Name,
		"default": project.Default,
	})
	if err != nil {
		return nil, fmt.Errorf("store project information: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
