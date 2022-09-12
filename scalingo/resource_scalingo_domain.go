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
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	domainName, _ := d.Get("common_name").(string)
	appID, _ := d.Get("app").(string)

	domain, err := client.DomainsAdd(appID, scalingo.Domain{
		Name: domainName,
	})
	if err != nil {
		return diag.Errorf("fail to add domain: %v", err)
	}
	d.SetId(domain.ID)

	return nil
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)

	domain, err := client.DomainsShow(appID, d.Id())
	if err != nil {
		return diag.Errorf("fail to get domain: %v", err)
	}

	d.SetId(domain.ID)
	err = d.Set("common_name", domain.Name)
	if err != nil {
		return diag.Errorf("fail to store domain name: %v", err)
	}

	return nil
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)

	err := client.DomainsRemove(appID, d.Id())
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
