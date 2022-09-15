package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v5"
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

// dataSourceScNotificationPlatformRead performs the Scalingo API lookup
func dataSourceScContainerSizeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	containers, err := client.ContainerSizesList(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	name, _ := d.Get("name").(string)

	i := 0
	for i < len(containers) && containers[i].Name != name {
		i++
	}
	if i >= len(containers) {
		return diag.Errorf("container '%v' not found", name)
	}

	d.SetId(containers[i].ID)
	err = SetAll(d, map[string]interface{}{
		"name":             containers[i].Name,
		"id":               containers[i].ID,
		"sku":              containers[i].SKU,
		"human_name":       containers[i].HumanName,
		"human_cpu":        containers[i].HumanCPU,
		"memory":           containers[i].Memory,
		"pids_limit":       containers[i].PidsLimit,
		"hourly_price":     containers[i].HourlyPrice,
		"thirtydays_price": containers[i].ThirtydaysPrice,
		"swap":             containers[i].Swap,
		"ordinal":          containers[i].Ordinal,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
