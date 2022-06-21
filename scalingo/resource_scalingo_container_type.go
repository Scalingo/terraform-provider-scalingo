package scalingo

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v4"
)

func resourceScalingoContainerType() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContainerTypeCreate,
		ReadContext:   resourceContainerTypeRead,
		UpdateContext: resourceContainerTypeUpdate,
		DeleteContext: resourceContainerTypeDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"amount": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceContainerTypeImport,
		},
	}
}

func resourceContainerTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	ctName, _ := d.Get("name").(string)

	resp, err := client.AppsScale(appID, &scalingo.AppsScaleParams{
		Containers: []scalingo.ContainerType{{
			Name:   ctName,
			Size:   d.Get("size").(string),
			Amount: d.Get("amount").(int),
		}},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	d.SetId(appID + ":" + ctName)

	return nil
}

func resourceContainerTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	ctName, _ := d.Get("name").(string)
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)
	d.SetId(appID + ":" + ctName)

	containers, err := client.AppsContainerTypes(appID)
	if err != nil {
		log.Printf("[INFO] Error getting current formation %#v", err)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Successfully fetched current formation")
	for _, ct := range containers {
		log.Printf("[DEBUG] Current container type: %s", ct.Name)

		if ctName == ct.Name {
			log.Printf("[DEBUG] Found container type in formation")
			err = SetAll(d, map[string]interface{}{
				"amount": ct.Amount,
				"size":   ct.Size,
			})
			if err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	return nil
}

func resourceContainerTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	ctName, _ := d.Get("name").(string)
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)

	resp, err := client.AppsScale(appID, &scalingo.AppsScaleParams{
		Containers: []scalingo.ContainerType{{
			Name:   ctName,
			Size:   d.Get("size").(string),
			Amount: d.Get("amount").(int),
		}},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] Scaled the application")

	return nil
}

func resourceContainerTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
func resourceContainerTypeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("ID should have the following format: <app ID>:<container type name>")
	}
	appID := ids[0]
	ctName := ids[1]
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Container type name: %s", ctName)

	client, _ := meta.(*scalingo.Client)
	containers, err := client.AppsContainerTypes(appID)
	if err != nil {
		return nil, err
	}

	for _, ct := range containers {
		log.Printf("[DEBUG] Current container type: %s", ct.Name)

		if ctName == ct.Name {
			log.Printf("[DEBUG] Found the container type to import")
			d.SetId(appID + ":" + ctName)
			err = SetAll(d, map[string]interface{}{
				"name":   ctName,
				"app":    appID,
				"amount": ct.Amount,
				"size":   ct.Size,
			})
			if err != nil {
				return nil, err
			}

			return []*schema.ResourceData{d}, nil
		}
	}

	log.Printf("[DEBUG] No container type found")
	return nil, errors.New("not found")
}
