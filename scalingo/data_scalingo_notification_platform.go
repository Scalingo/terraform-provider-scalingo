package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v4"
)

func dataSourceScNotificationPlatform() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScNotificationPlatformRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logo_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"available_event_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// dataSourceScNotificationPlatformRead performs the Scalingo API lookup
func dataSourceScNotificationPlatformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	platforms, err := client.NotificationPlatformsList()
	if err != nil {
		return diag.FromErr(err)
	}

	name, ok := d.Get("name").(string)
	if !ok || name == "" {
		return diag.Errorf("name attribute is mandatory")
	}

	var selected *scalingo.NotificationPlatform
	for _, p := range platforms {
		if p.Name == name {
			selected = p
			break
		}
	}

	if selected == nil {
		return diag.Errorf("notification platform '%v' not found", name)
	}

	d.SetId(selected.ID)
	err = SetAll(d, map[string]interface{}{
		"name":                selected.Name,
		"display_name":        selected.DisplayName,
		"description":         selected.Description,
		"logo_url":            selected.LogoURL,
		"available_event_ids": selected.AvailableEventIDs,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
