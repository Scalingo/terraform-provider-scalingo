package scalingo

import (
	"errors"

	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform/helper/schema"
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
	}
}

func resourceAddonCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	providerId := d.Get("provider_id").(string)
	planName := d.Get("plan").(string)
	appId := d.Get("app").(string)

	planId, err := addonPlanID(client, providerId, planName)
	if err != nil {
		return err
	}

	d.Set("plan_id", planId)

	addon, err := client.AddonProvision(appId, providerId, planId)
	if err != nil {
		return err
	}

	d.Set("resource_id", addon.Addon.ResourceID)
	d.SetId(addon.Addon.ID)
	return nil
}

func resourceAddonRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appId := d.Get("app").(string)

	addon, err := client.AddonShow(appId, d.Id())
	if err != nil {
		return err
	}

	d.Set("resource_id", addon.ResourceID)
	d.Set("plan", addon.Plan.Name)
	d.Set("plan_id", addon.PlanID)
	d.SetId(addon.ID)

	return nil
}

func resourceAddonUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appId := d.Get("app").(string)
	providerId := d.Get("provider_id").(string)

	d.Partial(true)

	if d.HasChange("plan") {
		planId, err := addonPlanID(client, providerId, d.Get("plan").(string))
		if err != nil {
			return err
		}

		res, err := client.AddonUpgrade(appId, d.Id(), planId)
		if err != nil {
			return err
		}

		d.Set("plan_id", res.Addon.PlanID)
		d.SetPartial("plan")
		d.SetPartial("plan_id")
	}

	d.Partial(false)
	return nil
}

func resourceAddonDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appId := d.Get("app").(string)

	err := client.AddonDestroy(appId, d.Id())
	if err != nil {
		return err
	}

	return nil
}

func addonPlanID(client *scalingo.Client, providerId, name string) (string, error) {
	plans, err := client.AddonProviderPlansList(providerId)
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
