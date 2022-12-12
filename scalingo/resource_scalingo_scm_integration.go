package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v6"
)

func resourceScalingoScmIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScmIntegrationCreate,
		ReadContext:   resourceScmIntegrationRead,
		DeleteContext: resourceScmIntegrationDelete,
		Description:   "Resource representing an SCM Integration are the link between an account and a source code management platform",

		Schema: map[string]*schema.Schema{
			"scm_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Type of integration (github-enterprise/gitlab-self-hosted), others should be created from dashboard",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "URL to the SCM Platform domain",
			},
			"access_token": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "API Access Token to communicate with the API (GitHub Enterprise and Gitlab Self-Hosted only)",
			},
			"uid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique ID of the SCM account",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username in the SCM account",
			},
			"avatar_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Avatar URL in the SCM account",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Email in the SCM account",
			},
			"profile_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Profile URL in the SCM account",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceScmIntegrationImport,
		},
	}
}

func resourceScmIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	scmType, _ := d.Get("scm_type").(string)
	url, _ := d.Get("url").(string)
	accessToken, _ := d.Get("access_token").(string)

	integration, err := client.SCMIntegrationsCreate(ctx, scalingo.SCMType(scmType), url, accessToken)
	if err != nil {
		return diag.Errorf("fail to create scm integration: %v", err)
	}
	d.SetId(integration.ID)

	return nil
}

func resourceScmIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	integration, err := client.SCMIntegrationsShow(ctx, id)
	if err != nil {
		return diag.Errorf("fail to fetch scm integration: %v", err)
	}
	err = SetAll(d, map[string]interface{}{
		"scm_type":    integration.SCMType,
		"url":         integration.URL,
		"uid":         integration.UID,
		"username":    integration.Username,
		"avatar_url":  integration.AvatarURL,
		"email":       integration.Email,
		"profile_url": integration.ProfileURL,
	})
	if err != nil {
		return diag.Errorf("fail to store scm integration information: %v", err)
	}
	d.SetId(integration.ID)

	return nil
}

func resourceScmIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	err := client.SCMIntegrationsDelete(ctx, id)
	if err != nil {
		return diag.Errorf("fail to delete scm integration: %v", err)
	}

	return nil
}

func resourceScmIntegrationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	diags := resourceScmIntegrationRead(ctx, d, meta)
	if err := DiagnosticError(diags); err != nil {
		return nil, fmt.Errorf("fail to read scm integration: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
