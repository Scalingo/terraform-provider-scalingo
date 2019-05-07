package scalingo

import (
	"errors"
	"log"
	"strings"

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
			State: resourceContainerTypeImport,
		},
	}
}

func resourceContainerTypeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)
	ctName := d.Get("name").(string)

	_, err := client.AppsScale(appID, &scalingo.AppsScaleParams{
		Containers: []scalingo.ContainerType{{
			Name:   ctName,
			Size:   d.Get("size").(string),
			Amount: d.Get("amount").(int),
		}},
	})
	if err != nil {
		return err
	}

	d.SetId(ctName)

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

// resourceContainerTypeImport is called when importing a new container_type
// resource. The ID must be "appID:containerTypeName" such as
// "5a155aa8f112e20010779b7a:web".
//
// Usage:
// $ terraform import terraform ID appID:containerTypeName
//
// Example:
// $ terraform import module.vpn-addon-service.module.app.module.app.scalingo_container_type.default 5a155aa8f112e20010779b7a:web
func resourceContainerTypeImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("address should have the following format: <app ID>:<container type name>")
	}
	appID := ids[0]
	ctName := ids[1]
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)

	d.SetId(ctName)
	d.Set("app", appID)

	resourceContainerTypeRead(d, meta)

	return []*schema.ResourceData{d}, nil
}
