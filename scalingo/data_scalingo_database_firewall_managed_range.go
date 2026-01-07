package scalingo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func dataSourceScDatabaseFirewallManagedRange() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScDatabaseFirewallManagedRangeRead,
		Description: "Database firewall managed IP range retrieved from the Database API",

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the database",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the managed firewall IP range to look up",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the managed firewall IP range",
			},
		},
	}
}

func dataSourceScDatabaseFirewallManagedRangeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	databaseID, _ := d.Get("database_id").(string)
	name, _ := d.Get("name").(string)

	appID, addonID, err := getDBAPIContext(ctx, client, databaseID)
	if err != nil {
		return diag.Errorf("resolve database context: %v", err)
	}

	managedRanges, err := previewClient.FirewallRulesGetManagedRanges(ctx, appID, addonID)
	if err != nil {
		return diag.Errorf("list firewall managed ranges: %v", err)
	}

	var selected *scalingo.FirewallManagedRange
	for _, managedRange := range managedRanges {
		if managedRange.Name == name {
			selected = &managedRange
			break
		}
	}

	if selected == nil {
		return diag.Errorf("managed range '%v' not found", name)
	}

	d.SetId(selected.ID)
	err = SetAll(d, map[string]interface{}{
		"database_id": databaseID,
		"name":        selected.Name,
		"id":          selected.ID,
	})
	if err != nil {
		return diag.Errorf("store managed range attributes: %v", err)
	}

	return nil
}
