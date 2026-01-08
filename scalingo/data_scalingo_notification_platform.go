package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v9"
)

func dataSourceScNotificationPlatform() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScNotificationPlatformRead,
		Description: "Notification platforms are the different destination to which notifications and alerts about an application can be sent",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug name of the notification platform",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human-enriched name of the notification platform",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Logo image URL representing the notification platform",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Textual description",
			},
			"available_event_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of event IDs which can be sent through this platform",
			},
		},
	}
}

// dataSourceScNotificationPlatformRead performs the Scalingo API lookup
func dataSourceScNotificationPlatformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	platforms, err := client.NotificationPlatformsList(ctx)
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
		return diag.Errorf("fail to store notification platform attributes: %v", err)
	}

	return nil
}
