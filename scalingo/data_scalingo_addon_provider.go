package scalingo

import (
	"context"

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
			"logo_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScAddonProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	name, ok := d.Get("name").(string)
	if !ok || name == "" {
		return diag.Errorf("name attribute is mandatory")
	}

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
	err = SetAll(d, map[string]interface{}{
		"id":       selected.ID,
		"name":     selected.Name,
		"logo_url": selected.LogoURL,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
