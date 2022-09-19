package scalingo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
)

func resourceScalingoAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddonCreate,
		ReadContext:   resourceAddonRead,
		UpdateContext: resourceAddonUpdate,
		DeleteContext: resourceAddonDelete,

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"plan": {
				Type:     schema.TypeString,
				Required: true,
			},
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"plan_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceAddonImport,
		},
	}
}

func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	providerID, _ := d.Get("provider_id").(string)
	planName, _ := d.Get("plan").(string)
	appID, _ := d.Get("app").(string)

	planID, err := addonPlanID(ctx, client, providerID, planName)
	if err != nil {
		return diag.Errorf("fail to get addon plan id: %v", err)
	}

	if err := d.Set("plan_id", planID); err != nil {
		return diag.FromErr(err)
	}

	res, err := client.AddonProvision(ctx, appID, scalingo.AddonProvisionParams{
		AddonProviderID: providerID,
		PlanID:          planID,
	})
	if err != nil {
		return diag.Errorf("fail to provision addon: %v", err)
	}

	err = waitUntilProvisionned(ctx, client, res.Addon)
	if err != nil {
		return diag.Errorf("fail to wait for the addon to be provisioned: %v", err)
	}

	d.SetId(res.Addon.ID)
	if err := d.Set("resource_id", res.Addon.ResourceID); err != nil {
		return diag.Errorf("fail to store addon resource id: %v", err)
	}

	return nil
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)

	addon, err := client.AddonShow(ctx, appID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.MarkNewResource()
			return nil
		}
		return diag.Errorf("fail to get addon details: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"resource_id": addon.ResourceID,
		"plan":        addon.Plan.Name,
		"plan_id":     addon.Plan.ID,
		"provider_id": addon.AddonProvider.ID,
	})
	if err != nil {
		return diag.Errorf("fail to store addon information: %v", err)
	}

	d.SetId(addon.ID)

	return nil
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	providerID, _ := d.Get("provider_id").(string)

	if d.HasChange("plan") {
		planID, err := addonPlanID(ctx, client, providerID, d.Get("plan").(string))
		if err != nil {
			return diag.Errorf("fail to get addon plan id: %v", err)
		}

		res, err := client.AddonUpgrade(ctx, appID, d.Id(), scalingo.AddonUpgradeParams{
			PlanID: planID,
		})
		if err != nil {
			return diag.Errorf("fail to upgrade addon: %v", err)
		}

		err = waitUntilProvisionned(ctx, client, res.Addon)
		if err != nil {
			return diag.Errorf("fail to wait for the addon to be provisioned: %v", err)
		}

		if err := d.Set("plan_id", res.Addon.Plan.ID); err != nil {
			return diag.Errorf("fail to store addon plan id: %v", err)
		}
	}

	return nil
}

func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)

	err := client.AddonDestroy(ctx, appID, d.Id())
	if err != nil {
		return diag.Errorf("fail to destroy addon: %v", err)
	}

	return nil
}

func addonPlanID(ctx context.Context, client *scalingo.Client, providerID, name string) (string, error) {
	plans, err := client.AddonProviderPlansList(ctx, providerID)
	if err != nil {
		return "", err
	}

	planList := ""
	for _, plan := range plans {
		if plan.Name == name {
			return plan.ID, nil
		}

		planList += ", " + plan.Name
	}

	return "", fmt.Errorf("Invalid plan name, possible values are: " + planList)
}

func resourceAddonImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, fmt.Errorf("address should have the following format: <appid>:<addonid>")
	}

	d.SetId(ids[1])
	if err := d.Set("app", ids[0]); err != nil {
		return nil, err
	}

	diags := resourceAddonRead(ctx, d, meta)
	err := DiagnosticError(diags)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func waitUntilProvisionned(ctx context.Context, client *scalingo.Client, addon scalingo.Addon) error {
	var err error
	timer := time.NewTimer(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	for addon.Status != scalingo.AddonStatusRunning {
		addon, err = client.AddonShow(ctx, addon.AppID, addon.ID)
		if err != nil {
			return err
		}
		// Do not wait for next tick to get out of the loop
		if addon.Status == scalingo.AddonStatusRunning {
			return nil
		}
		select {
		case <-timer.C:
			return fmt.Errorf("addon provisioning timed out")
		case <-ticker.C:
		}
	}
	return nil
}
