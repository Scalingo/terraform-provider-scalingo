package scalingo

import (
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo"
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

	d.SetId(appID + ":" + ctName)

	return nil
}

func resourceContainerTypeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)
	ctName := d.Get("name").(string)
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)
	d.SetId(appID + ":" + ctName)

	containers, err := client.AppsPs(appID)
	if err != nil {
		log.Printf("[INFO] Error getting current formation %#v", err)
		return err
	}

	log.Printf("[INFO] Successfuly fetched current formation")
	for _, ct := range containers {
		log.Printf("[DEBUG] Current container type: %s", ct.Name)

		if ctName == ct.Name {
			log.Printf("[DEBUG] Found container type in formation")
			d.Set("amount", ct.Amount)
			d.Set("size", ct.Size)
			break
		}
	}

	return nil
}

func resourceContainerTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)
	ctName := d.Get("name").(string)
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)

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

	log.Printf("[DEBUG] Scaled the application")

	return nil
}

func resourceContainerTypeDelete(d *schema.ResourceData, meta interface{}) error {
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
		return nil, errors.New("ID should have the following format: <app ID>:<container type name>")
	}
	appID := ids[0]
	ctName := ids[1]
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)

	client := meta.(*scalingo.Client)
	containers, err := client.AppsPs(appID)
	if err != nil {
		return nil, err
	}

	for _, ct := range containers {
		log.Printf("[DEBUG] Current container type: %s", ct.Name)

		if ctName == ct.Name {
			log.Printf("[DEBUG] Found the container type to import")
			d.SetId(appID + ":" + ctName)
			d.Set("name", ctName)
			d.Set("app", appID)
			d.Set("amount", ct.Amount)
			d.Set("size", ct.Size)

			return []*schema.ResourceData{d}, nil
		}
	}

	log.Printf("[DEBUG] No container type found")
	return nil, errors.New("not found")
}
