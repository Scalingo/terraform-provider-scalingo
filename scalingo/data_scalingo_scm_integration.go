package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
)

func dataSourceScScmIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScScmIntegrationRead,

		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scm_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScScmIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	scmType := d.Get("scm_type").(string)
	url, _ := d.Get("url").(string)

	integrations, err := client.SCMIntegrationsList(ctx)
	if err != nil {
		return diag.Errorf("fail to fetch integrations: %v", err)
	}

	selectedIntegrations := keepIf(integrations, func(integration scalingo.SCMIntegration) bool {
		selected := true
		if scmType != "" {
			selected = selected && (scalingo.SCMType(scmType) == integration.SCMType)
		}
		if url != "" {
			selected = selected && (url == integration.URL)
		}
		return selected
	})

	if len(selectedIntegrations) != 1 {
		return diag.Errorf("fail to find the selected integration")
	}

	integration := selectedIntegrations[0]
	err = SetAll(d, map[string]interface{}{
		"scm_type":    integration.SCMType,
		"url":         integration.URL,
		"uid":         integration.UID,
		"username":    integration.Username,
		"avatar_url":  integration.AvatarURL,
		"email":       integration.Email,
		"profile_url": integration.ProfileURL,
		"owner_id":    integration.Owner.ID,
	})
	if err != nil {
		return diag.Errorf("fail to store scm integration information: %v", err)
	}
	d.SetId(integration.ID)

	return nil
}
