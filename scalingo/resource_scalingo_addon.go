package scalingo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func resourceScalingoAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddonCreate,
		ReadContext:   resourceAddonRead,
		UpdateContext: resourceAddonUpdate,
		DeleteContext: resourceAddonDelete,
		Description:   "Resource representing an Addon attached to an Application based on an AddonProvider",

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of slug name of the addon provider",
			},
			"plan": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the plan of the addon to provision",
			},
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the application which will receive the addon",
			},
			"plan_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the plan which was provisioned",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human readable ID of the addon which is provisioned",
			},
			"database_features": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of enabled features for the addon (Database addons only)",
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

	err = waitUntilProvisioned(ctx, client, res.Addon)
	if err != nil {
		return diag.Errorf("fail to wait for the addon to be provisioned: %v", err)
	}

	d.SetId(res.Addon.ID)
	if err := d.Set("resource_id", res.Addon.ResourceID); err != nil {
		return diag.Errorf("fail to store addon resource id: %v", err)
	}

	databaseFeatures, _ := d.Get("database_features").([]interface{})
	for _, feature := range databaseFeatures {
		featureStr, _ := feature.(string)
		_, err := client.DatabaseEnableFeature(ctx, appID, res.Addon.ID, featureStr)
		if err != nil {
			return diag.Errorf("fail to add feature on database addon id: %v", err)
		}
		err = waitUntilDatabaseFeatureActivated(ctx, client, res.Addon, featureStr)
		if err != nil {
			return diag.Errorf("fail to activate feature on database addon id: %v", err)
		}
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

	providers, err := client.AddonProvidersList(ctx)
	if err != nil {
		return diag.Errorf("fail to list addon providers: %v", err)
	}
	features := []string{}
	if addonIsDatabase(providers, addon) {
		db, err := client.DatabaseShow(ctx, addon.AppID, addon.ID)
		if err != nil {
			return diag.Errorf("fail to get database metadata for addon %v: %v", addon.ID, err)
		}
		for _, feature := range db.Features {
			features = append(features, feature.Name)
		}
		err = d.Set("database_features", features)
		if err != nil {
			return diag.Errorf("fail to set database_features parameter: %v", err)
		}
	}

	d.SetId(addon.ID)

	return nil
}

func addonIsDatabase(providers []*scalingo.AddonProvider, addon scalingo.Addon) bool {
	addonProviders := keepIf(providers, func(p *scalingo.AddonProvider) bool {
		return p.ID == addon.AddonProvider.ID
	})
	if len(addonProviders) == 0 {
		return false
	}

	return strings.HasPrefix(strings.ToLower(addonProviders[0].Category.Name), "database")
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	providerID, _ := d.Get("provider_id").(string)

	addon, err := client.AddonShow(ctx, appID, d.Id())
	if err != nil {
		return diag.Errorf("fail to get addon information for %v: %v", d.Id(), err)
	}

	if d.HasChange("plan") {
		planID, err := addonPlanID(ctx, client, providerID, d.Get("plan").(string))
		if err != nil {
			return diag.Errorf("fail to get addon plan id: %v", err)
		}

		res, err := client.AddonUpgrade(ctx, addon.AppID, addon.ID, scalingo.AddonUpgradeParams{
			PlanID: planID,
		})
		if err != nil {
			return diag.Errorf("fail to upgrade addon: %v", err)
		}

		err = waitUntilProvisioned(ctx, client, res.Addon)
		if err != nil {
			return diag.Errorf("fail to wait for the addon to be provisioned: %v", err)
		}

		if err := d.Set("plan_id", res.Addon.Plan.ID); err != nil {
			return diag.Errorf("fail to store addon plan id: %v", err)
		}
		addon = res.Addon
	}

	if d.HasChange("database_features") {
		db, err := client.DatabaseShow(ctx, addon.AppID, addon.ID)
		if err != nil {
			return diag.Errorf("fail to get database metadata from addon %v: %v", addon.ID, err)
		}
		databaseFeatures, _ := d.Get("database_features").([]interface{})

		err = compareAndApplyDatabaseFeatures(ctx, client, addon, db, databaseFeatures)
		if err != nil {
			return diag.Errorf("fail to compare and apply database features of %v: %v", addon.ID, err)
		}
	}

	return nil
}

func compareAndApplyDatabaseFeatures(ctx context.Context, client *scalingo.Client, addon scalingo.Addon, db scalingo.Database, databaseFeatures []interface{}) error {
	featuresToAdd := []string{}
	featuresToRemove := []string{}

	for _, feature := range databaseFeatures {
		toAdd := true
		for _, dbFeature := range db.Features {
			if dbFeature.Name == feature.(string) {
				toAdd = false
			}
		}
		if toAdd {
			featuresToAdd = append(featuresToAdd, feature.(string))
		}
	}

	for _, dbFeature := range db.Features {
		toRemove := true
		for _, feature := range databaseFeatures {
			if dbFeature.Name == feature.(string) {
				toRemove = false
			}
		}
		if toRemove {
			featuresToRemove = append(featuresToRemove, dbFeature.Name)
		}
	}

	for _, feature := range featuresToAdd {
		_, err := client.DatabaseEnableFeature(ctx, addon.AppID, addon.ID, feature)
		if err != nil {
			return fmt.Errorf("fail to enable database feature for addon %v: %v", addon.ID, feature)
		}
		err = waitUntilDatabaseFeatureActivated(ctx, client, addon, feature)
		if err != nil {
			return fmt.Errorf("fail to wait until database feature '%v' is enabled %v: %v", feature, addon.ID, err)
		}
	}

	for _, feature := range featuresToRemove {
		_, err := client.DatabaseDisableFeature(ctx, addon.AppID, addon.ID, feature)
		if err != nil {
			return fmt.Errorf("fail to disable feature '%v' for %v: %v", feature, addon.ID, err)
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

	return "", fmt.Errorf("Invalid plan name, possible values are: %s", planList)
}

func resourceAddonImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, fmt.Errorf("address should have the following format: <appID>:<addonID>")
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

func waitUntilProvisioned(ctx context.Context, client *scalingo.Client, addon scalingo.Addon) error {
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

func waitUntilDatabaseFeatureActivated(ctx context.Context, client *scalingo.Client, addon scalingo.Addon, feature string) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		db, err := client.DatabaseShow(ctx, addon.AppID, addon.ID)
		if err != nil {
			return fmt.Errorf("fail to refresh database metadata: %w", err)
		}
		for _, f := range db.Features {
			if f.Name == feature && f.Status != scalingo.DatabaseFeatureStatusPending {
				switch f.Status {
				case scalingo.DatabaseFeatureStatusActivated:
					return nil
				case scalingo.DatabaseFeatureStatusFailed:
					return fmt.Errorf("fail to enable feature %v, please contact support: %w", feature, err)
				}
			}
		}
	}
	return nil
}
