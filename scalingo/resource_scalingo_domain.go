package scalingo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
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

		Schema: map[string]*schema.Schema{
			"common_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"canonical": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	domainName, _ := d.Get("common_name").(string)
	canonical, _ := d.Get("canonical").(bool)

	domain, err := client.DomainsAdd(ctx, appID, scalingo.Domain{
		Name:      domainName,
		Canonical: canonical,
	})
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

	// TODO add update when go scalingo is up to date

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
