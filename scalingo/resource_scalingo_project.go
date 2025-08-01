package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func resourceScalingoProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,

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
				Required:    false,
				Default:     false,
				Description: "Whether or not the project is a default project",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the project",
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	project, err := client.ProjectAdd(ctx, scalingo.ProjectAddParams{
		Name:    "",
		Default: false,
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
