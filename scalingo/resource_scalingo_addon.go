package scalingo

import (
	"errors"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo"
)

func resourceScalingoAddon() *schema.Resource {
	return &schema.Resource{
		Create: resourceAddonCreate,
		Read:   resourceAddonRead,
		Update: resourceAddonUpdate,
		Delete: resourceAddonDelete,

		Schema: map[string]*schema.Schema{
			"provider_id": &schema.Schema{
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
			State: resourceAddonImport,
		},
	}
}

func resourceAddonCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	providerID := d.Get("provider_id").(string)
	planName := d.Get("plan").(string)
	appID := d.Get("app").(string)

	planID, err := addonPlanID(client, providerID, planName)
	if err != nil {
		return err
	}

	d.Set("plan_id", planID)

	res, err := client.AddonProvision(appID, scalingo.AddonProvisionParams{
		AddonProviderID: providerID,
		PlanID:          planID,
	})
	if err != nil {
		return err
	}

	err = waitUntilProvisionned(client, res.Addon)
	if err != nil {
		return err
	}

	d.Set("resource_id", res.Addon.ResourceID)
	d.SetId(res.Addon.ID)
	return nil
}

func resourceAddonRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)

	addon, err := client.AddonShow(appID, d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.MarkNewResource()
			return nil
		}
		return err
	}

	d.Set("resource_id", addon.ResourceID)
	d.Set("plan", addon.Plan.Name)
	d.Set("plan_id", addon.Plan.ID)
	d.Set("provider_id", addon.AddonProvider.ID)
	d.SetId(addon.ID)

	return nil
}

func resourceAddonUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)
	providerID := d.Get("provider_id").(string)

	if d.HasChange("plan") {
		planID, err := addonPlanID(client, providerID, d.Get("plan").(string))
		if err != nil {
			return err
		}

		res, err := client.AddonUpgrade(appID, d.Id(), scalingo.AddonUpgradeParams{
			PlanID: planID,
		})
		if err != nil {
			return err
		}

		err = waitUntilProvisionned(client, res.Addon)
		if err != nil {
			return err
		}

		d.Set("plan_id", res.Addon.Plan.ID)
	}

	return nil
}

func resourceAddonDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)

	err := client.AddonDestroy(appID, d.Id())
	if err != nil {
		return err
	}

	return nil
}

func addonPlanID(client *scalingo.Client, providerID, name string) (string, error) {
	plans, err := client.AddonProviderPlansList(providerID)
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

	return "", errors.New("Invalid plan name, possible values are: " + planList)
}

func resourceAddonImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("address should have the following format: <appid>:<addonid>")
	}

	d.SetId(ids[1])
	d.Set("app", ids[0])

	resourceAddonRead(d, meta)

	return []*schema.ResourceData{d}, nil
}

func waitUntilProvisionned(client *scalingo.Client, addon scalingo.Addon) error {
	var err error
	timer := time.NewTimer(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	for addon.Status != scalingo.AddonStatusRunning {
		addon, err = client.AddonShow(addon.AppID, addon.ID)
		if err != nil {
			return err
		}
		// Do not wait for next tick to get out of the loop
		if addon.Status == scalingo.AddonStatusRunning {
			return nil
		}
		select {
		case <-timer.C:
			return errors.New("addon provisioning timed out")
		case <-ticker.C:
		}
	}
	return nil
}
