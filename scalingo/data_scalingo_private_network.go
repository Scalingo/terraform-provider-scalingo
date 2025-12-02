package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func dataSourceScPrivateNetworkDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScPrivateNetworkDomainsRead,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainImporter,
		},
		Description: "Resource representing a the private network domains of an application",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application",
			},
			"page": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Page number to retrieve",
			},
			"page_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     50,
				Description: "Number of items per page (max 50)",
			},
		},
	}
}

func dataSourceScPrivateNetworkDomainsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	page, _ := d.Get("page").(uint)
	pageSize, _ := d.Get("page_size").(uint)

	domains, err := client.PrivateNetworksDomainsList(ctx, appID, page, pageSize)
	if err != nil {
		return diag.Errorf("fail to list project private network domains: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"domains": domains,
	})
	if err != nil {
		return diag.Errorf("fail to store project private network domains list: %v", err)
	}
	d.SetId(appID)

	return nil
}
