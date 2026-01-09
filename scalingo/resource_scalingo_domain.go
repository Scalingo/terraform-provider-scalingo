package scalingo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v9"
)

func resourceScalingoDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		UpdateContext: resourceDomainUpdate,
		ReadContext:   resourceDomainRead,
		DeleteContext: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainImporter,
		},
		Description: "Resource representing a custom domain targeting an application",

		Schema: map[string]*schema.Schema{
			"common_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Common Name (hostname) of the DNS entry which will target the application",
			},
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application",
			},
			"canonical": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "If true, all requests will be redirected to this domain (one per application)",
			},
			"letsencrypt_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true (default), the domain will be secured with a Let's Encrypt certificate",
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	domainName, _ := d.Get("common_name").(string)
	canonical, _ := d.Get("canonical").(bool)
	letsEncryptEnabled, _ := d.Get("letsencrypt_enabled").(bool)

	params := scalingo.DomainsAddParams{
		Name:               domainName,
		Canonical:          &canonical,
		LetsEncryptEnabled: &letsEncryptEnabled,
	}
	domain, err := client.DomainsAdd(ctx, appID, params)
	if err != nil {
		return diag.Errorf("fail to add domain: %v", err)
	}
	d.SetId(domain.ID)

	return nil
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	canonical, _ := d.Get("canonical").(bool)

	if d.HasChange("canonical") {
		var domain scalingo.Domain
		var err error
		if canonical {
			domain, err = client.DomainSetCanonical(ctx, appID, d.Id())
		} else {
			// This may cause an issue when the user is changing which domain is canonical
			// This may add the new canonical flag first and remove the old one after
			// In this case the newly set canonical flag will be removed instead of the old one
			domain, err = client.DomainUnsetCanonical(ctx, appID)
		}
		if err != nil {
			return diag.Errorf("fail to update domain: %v", err)
		}
		err = d.Set("canonical", domain.Canonical)
		if err != nil {
			return diag.Errorf("fail to store domain information: %v", err)
		}
	}
	return nil
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	domain, err := client.DomainsShow(ctx, appID, d.Id())
	if err != nil {
		return diag.Errorf("fail to get domain: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"common_name": domain.Name,
		"canonical":   domain.Canonical,
	})
	if err != nil {
		return diag.Errorf("fail to store domain information: %v", err)
	}
	d.SetId(domain.ID)

	return nil
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)

	err := client.DomainsRemove(ctx, appID, d.Id())
	if err != nil {
		return diag.Errorf("fail to remove domain: %v", err)
	}

	return nil
}

func resourceDomainImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if !strings.Contains(d.Id(), ":") {
		return nil, fmt.Errorf("schema must be app_id:domain_id")
	}
	split := strings.Split(d.Id(), ":")
	d.SetId(split[1])

	err := d.Set("app", split[0])
	if err != nil {
		return nil, fmt.Errorf("fail to set app name: %v", err)
	}

	diags := resourceDomainRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to read domain: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
