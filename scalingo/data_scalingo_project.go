package scalingo

import (
	"context"
	"fmt"

	"github.com/Scalingo/go-scalingo/v8"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScProjectRead,
		Description: "Data source representing a project",

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
	}
}

func dataSourceScProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	projectID, ok := d.Get("id").(string)
	if !ok || projectID == "" {
		return diag.Errorf("id attribute is mandatory")
	}

	project, err := client.ProjectGet(ctx, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)
	err = SetAll(d, map[string]interface{}{
		"name":    project.Name,
		"default": project.Default,
	})
	if err != nil {
		return diag.Errorf("store project information: %v", err)
	}
	tflog.Info(ctx, fmt.Sprintf("Fetched project '%s' with ID %s", project.Name, projectID))

	return nil
}
