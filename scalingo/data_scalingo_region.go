package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v5"
)

func dataSourceScRegion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScRegionsRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_api": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	regions, err := client.RegionsList(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	name, _ := d.Get("name").(string)

	var selected scalingo.Region
	for _, v := range regions {
		if v.Name == name {
			selected = v
			break
		}
	}

	if selected == (scalingo.Region{}) {
		return diag.Errorf("region '%v' not found", name)
	}

	d.SetId(selected.Name)
	err = SetAll(d, map[string]interface{}{
		"name":         selected.Name,
		"display_name": selected.DisplayName,
		"api":          selected.API,
		"dashboard":    selected.Dashboard,
		"database_api": selected.DatabaseAPI,
		"ssh":          selected.SSH,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
