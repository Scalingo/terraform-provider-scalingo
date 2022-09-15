package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v5"
)

func dataSourceScAddonProvider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScAddonProviderRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plans": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"position": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"on_demand": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"disabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"disabled_alternative_plan_id": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sku": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hds_available": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func dataSourceScAddonProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	name, _ := d.Get("name").(string)

	addonProviders, err := client.AddonProvidersList(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var selected *scalingo.AddonProvider

	for _, v := range addonProviders {
		if v.Name == name {
			selected = v
			break
		}
	}

	if selected == nil {
		return diag.Errorf("addon provider '%v' not found", name)
	}
	d.SetId(selected.ID)
	categoryMap := map[string]interface{}{
		"id":          selected.Category.ID,
		"name":        selected.Category.Name,
		"description": selected.Category.Description,
		"position":    fmt.Sprintf("%d", selected.Category.Position),
	}

	planMap := make([]map[string]interface{}, 0)
	for _, v := range selected.Plans {
		planMap = append(planMap, map[string]interface{}{
			"id":                           v.ID,
			"name":                         v.Name,
			"display_name":                 v.DisplayName,
			"price":                        v.Price,
			"description":                  v.Description,
			"position":                     v.Position,
			"on_demand":                    v.OnDemand,
			"disabled":                     v.Disabled,
			"disabled_alternative_plan_id": v.DisabledAlternativePlanID,
			"sku":                          v.SKU,
			"hds_available":                v.HDSAvailable,
		})
	}

	err = SetAll(d, map[string]interface{}{
		"id":                selected.ID,
		"name":              selected.Name,
		"short_description": selected.ShortDescription,
		"description":       selected.Description,
		"category":          categoryMap,
		"provider_name":     selected.ProviderName,
		"provider_url":      selected.ProviderURL,
		"plans":             planMap,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
