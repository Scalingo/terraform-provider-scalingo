package scalingo

import (
	scalingo "github.com/Scalingo/go-scalingo"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScalingoContainerType() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerTypeCreate,
		Read:   resourceContainerTypeRead,
		Update: resourceContainerTypeUpdate,
		Delete: resourceContainerTypeDelete,

		Schema: map[string]*schema.Schema{
			"app": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"amount": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceContainerTypeCreate(d *schema.ResourceData, meta interface{}) error {
	/*
		client := meta.(*scalingo.Client)

		appID := d.Get("app").(string)
		client.AppsScale(appID, &scalingo.AppsScaleParams{})

			collaborator, err := client.ContainerTypeAdd(d.Get("app").(string), d.Get("email").(string))
			if err != nil {
				return err
			}

			d.Set("username", collaborator.Username)
			d.Set("status", collaborator.Status)

			d.SetId(collaborator.ID)
	*/

	return nil
}

func resourceContainerTypeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)
	ctName := d.Get("name").(string)
	d.SetId(ctName)

	containers, err := client.AppsPs(appID)
	if err != nil {
		return err
	}
	for _, ct := range containers {
		if ctName == ct.Name {
			d.Set("amount", ct.Amount)
			d.Set("size", ct.Size)
			break
		}
	}

	return nil
}

func resourceContainerTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	/*
		client := meta.(*scalingo.Client)
	*/

	return nil
}

func resourceContainerTypeDelete(d *schema.ResourceData, meta interface{}) error {
	/*
		client := meta.(*scalingo.Client)

		err := client.ContainerTypeRemove(d.Get("app").(string), d.Id())
		if err != nil {
			return err
		}
	*/

	return nil
}
