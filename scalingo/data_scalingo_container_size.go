package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func dataSourceScContainerSize() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScContainerSizeRead,
		Description: "Container Sizes represents the definitions of the size of container types as well as their limits",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug name of the container size",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the container size",
			},
			"sku": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalogue reference of the container size",
			},
			"human_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human-enriched name of the container size",
			},
			"human_cpu": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human readable description of CPU quota",
			},
			"memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Container memory limit (in bytes)",
			},
			"swap": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Container swap memory limit (in bytes)",
			},
			"pids_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of processes (Process IDentifiers)",
			},
			"ordinal": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Position for editorial sorting of container size",
			},
		},
	}
}

func dataSourceScContainerSizeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	containers, err := client.ContainerSizesList(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	name, _ := d.Get("name").(string)

	var containerSize scalingo.ContainerSize
	for _, v := range containers {
		if v.Name == name {
			containerSize = v
			break
		}
	}
	if containerSize.ID == "" {
		return diag.Errorf("container size '%v' not found", name)
	}

	d.SetId(containerSize.ID)
	err = SetAll(d, map[string]interface{}{
		"name":       containerSize.Name,
		"id":         containerSize.ID,
		"sku":        containerSize.SKU,
		"human_name": containerSize.HumanName,
		"human_cpu":  containerSize.HumanCPU,
		"memory":     containerSize.Memory,
		"pids_limit": containerSize.PidsLimit,
		"swap":       containerSize.Swap,
		"ordinal":    containerSize.Ordinal,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
