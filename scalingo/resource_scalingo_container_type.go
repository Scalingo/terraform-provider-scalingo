package scalingo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v6"
)

func resourceScalingoContainerType() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContainerTypeCreate,
		ReadContext:   resourceContainerTypeRead,
		UpdateContext: resourceContainerTypeUpdate,
		DeleteContext: resourceContainerTypeDelete,
		Description:   "Resource representing a container type, allowing to scale an application containers",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the container type",
			},
			"amount": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Number of containers to boot for this type",
			},
			"size": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Size of the container (S/M/L/etc.)",
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

	resp, err := client.AppsScale(ctx, appID, &scalingo.AppsScaleParams{
		Containers: []scalingo.ContainerType{{
			Name:   ctName,
			Size:   d.Get("size").(string),
			Amount: d.Get("amount").(int),
		}},
	})
	if err != nil {
		return diag.Errorf("fail to scale application: %v", err)
	}
	defer resp.Body.Close()

	d.SetId(appID + ":" + ctName)

	return nil
}

func resourceContainerTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	ctName, _ := d.Get("name").(string)
	d.SetId(appID + ":" + ctName)

	containers, err := client.AppsContainerTypes(ctx, appID)
	if err != nil {
		return diag.Errorf("fail to list container types: %v", err)
	}

	for _, ct := range containers {
		if ctName == ct.Name {
			err = SetAll(d, map[string]interface{}{
				"amount": ct.Amount,
				"size":   ct.Size,
			})
			if err != nil {
				return diag.Errorf("fail to store container type information: %v", err)
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

	resp, err := client.AppsScale(ctx, appID, &scalingo.AppsScaleParams{
		Containers: []scalingo.ContainerType{{
			Name:   ctName,
			Size:   d.Get("size").(string),
			Amount: d.Get("amount").(int),
		}},
	})
	if err != nil {
		return diag.Errorf("fail to scale application: %v", err)
	}
	defer resp.Body.Close()

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
		return nil, fmt.Errorf("ID should have the following format: <app ID>:<container type name>")
	}
	appID := ids[0]
	ctName := ids[1]

	client, _ := meta.(*scalingo.Client)
	containers, err := client.AppsContainerTypes(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("fail to list container types: %v", err)
	}

	for _, ct := range containers {
		if ctName == ct.Name {
			d.SetId(appID + ":" + ctName)
			err = SetAll(d, map[string]interface{}{
				"name":   ctName,
				"app":    appID,
				"amount": ct.Amount,
				"size":   ct.Size,
			})
			if err != nil {
				return nil, fmt.Errorf("fail to store container information: %v", err)
			}

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("not found")
}
