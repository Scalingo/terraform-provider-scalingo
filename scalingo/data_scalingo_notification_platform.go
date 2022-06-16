package scalingo

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v4"
)

func dataSourceScNotificationPlatform() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceScNotificationPlatformRead,

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
func dataSourceScNotificationPlatformRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	platforms, err := client.NotificationPlatformsList()
	if err != nil {
		return fmt.Errorf("fail to read notification platforms: %v", err)
	}

	name, ok := d.Get("name").(string)
	if !ok || name == "" {
		return errors.New("name attribute is mandatory")
	}

	var selected *scalingo.NotificationPlatform
	for _, p := range platforms {
		if p.Name == name {
			selected = p
			break
		}
	}

	if selected == nil {
		return fmt.Errorf("notification platform '%v' not found", name)
	}

	d.SetId(selected.ID)
	d.Set("name", selected.Name)
	d.Set("display_name", selected.DisplayName)
	d.Set("description", selected.Description)
	d.Set("logo_url", selected.LogoURL)
	d.Set("available_event_ids", selected.AvailableEventIDs)

	return nil
}
