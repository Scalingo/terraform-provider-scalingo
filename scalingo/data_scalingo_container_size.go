package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v6"
)

func dataSourceScContainerSize() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScContainerSizeRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sku": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"human_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"human_cpu": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pids_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hourly_price": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"thirtydays_price": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"swap": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ordinal": {
				Type:     schema.TypeInt,
				Computed: true,
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
		"name":             containerSize.Name,
		"id":               containerSize.ID,
		"sku":              containerSize.SKU,
		"human_name":       containerSize.HumanName,
		"human_cpu":        containerSize.HumanCPU,
		"memory":           containerSize.Memory,
		"pids_limit":       containerSize.PidsLimit,
		"hourly_price":     containerSize.HourlyPrice,
		"thirtydays_price": containerSize.ThirtydaysPrice,
		"swap":             containerSize.Swap,
		"ordinal":          containerSize.Ordinal,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
