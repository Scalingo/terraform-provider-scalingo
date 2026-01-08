package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v9"
)

func dataSourceScAddonProvider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScAddonProviderRead,
		Description: "Addon Providers include metadata for all addons provided on the platform",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the addon provider",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the addon provider",
			},
			"short_description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One-line textual description",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Markdown-formatted multi-line description",
			},
			"category": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "Name of the addon category",
			},
			"provider_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the entity providing the addon",
			},
			"provider_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Homepage of the entity providing the addon",
			},
			"plans": {
				Type:        schema.TypeList,
				Description: "List of available plans for this provider",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the plan",
						},
						"display_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human readable name (multiword)",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Slug name of the plan (underscore case)",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Markdown-formatted multiline description",
						},
						"position": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Position for editorial sorting of addon plans",
						},
						"on_demand": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Can the plan be provisioned automatically or should it be validated by Scalingo support",
						},
						"disabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Can the plan be provisioned?",
						},
						"disabled_alternative_plan_id": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Filled if disabled: alternative plan to provision instead",
						},
						"sku": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Catalogue reference of the plan",
						},
						"hds_available": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Is the plan available for Health Data Hosting applications?",
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
	category := map[string]interface{}{
		"id":          selected.Category.ID,
		"name":        selected.Category.Name,
		"description": selected.Category.Description,
		"position":    fmt.Sprintf("%d", selected.Category.Position),
	}

	plans := []map[string]interface{}{}
	for _, v := range selected.Plans {
		plans = append(plans, map[string]interface{}{
			"id":                           v.ID,
			"name":                         v.Name,
			"display_name":                 v.DisplayName,
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
		"category":          category,
		"provider_name":     selected.ProviderName,
		"provider_url":      selected.ProviderURL,
		"plans":             plans,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
