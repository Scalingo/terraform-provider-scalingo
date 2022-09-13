package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
)

func resourceScalingoScmIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScmIntegrationCreate,
		ReadContext:   resourceScmIntegrationRead,
		DeleteContext: resourceScmIntegrationDelete,

		Schema: map[string]*schema.Schema{
			"scm_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"access_token": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
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
		return diag.Errorf("fail to add scm integration: %v", err)
	}
	d.SetId(integration.ID)

	return nil
}

func resourceScmIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	integration, err := client.SCMIntegrationsShow(ctx, id)
	if err != nil {
		return diag.Errorf("fail to fetch integration: %v", err)
	}
	err = SetAll(d, map[string]interface{}{
		"scm_type":    integration.SCMType,
		"url":         integration.URL,
		"uid":         integration.Uid,
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
	d.SetId(d.Id())

	diags := resourceGithubLinkRead(ctx, d, meta)
	if err := DiagnosticError(diags); err != nil {
		return nil, fmt.Errorf("fail to read scm integration: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
