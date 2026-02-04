package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v9"
)

func dataSourceScPrivateNetworkDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScPrivateNetworkDomainsRead,
		Description: "Resource representing a the private network domains of an application",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application",
			},
			"domains": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of private network domains attached to the application",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
				Description: "Number of items per page",
			},
		},
	}
}

func dataSourceScPrivateNetworkDomainsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, ok := d.Get("app").(string)
	if !ok || appID == "" {
		return diag.Errorf("app ID must be provided")
	}
	page, _ := d.Get("page").(int)
	pageSize, _ := d.Get("page_size").(int)

	domains, err := client.PrivateNetworksDomainsList(ctx, appID, uint(page), uint(pageSize))
	if err != nil {
		return diag.Errorf("list project private network domains: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"domains": domains.Data,
	})
	if err != nil {
		return diag.Errorf("store project private network domains list: %v", err)
	}
	d.SetId(appID)

	return nil
}
