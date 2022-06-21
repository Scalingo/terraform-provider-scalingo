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

func resourceScalingoAutoscaler() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAutoscalerCreate,
		ReadContext:   resourceAutoscalerRead,
		UpdateContext: resourceAutoscalerUpdate,
		DeleteContext: resourceAutoscalerDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"min_containers": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"max_containers": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"metric": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceAutoscalerImport,
		},
	}
}

func resourceAutoscalerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	log.Printf("[DEBUG] Application ID: %s", appID)

	autoscaler, err := client.AutoscalerAdd(appID, scalingo.AutoscalerAddParams{
		ContainerType: d.Get("container_type").(string),
		Metric:        d.Get("metric").(string),
		Target:        d.Get("target").(float64),
		MinContainers: d.Get("min_containers").(int),
		MaxContainers: d.Get("max_containers").(int),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(autoscaler.ID)

	return nil
}

func resourceAutoscalerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	appID, _ := d.Get("app").(string)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)
	log.Printf("[DEBUG] Application ID: %s", appID)

	autoscaler, err := client.AutoscalerShow(appID, id)
	if err != nil {
		log.Printf("[INFO] Error getting autoscaler %#v", err)
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Successfully fetched autoscaler")

	err = SetAll(d, map[string]interface{}{
		"container_type": autoscaler.ContainerType,
		"min_containers": autoscaler.MinContainers,
		"max_containers": autoscaler.MaxContainers,
		"metric":         autoscaler.Metric,
		"target":         autoscaler.Target,
		"disabled":       autoscaler.Disabled,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAutoscalerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	appID, _ := d.Get("app").(string)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)
	log.Printf("[DEBUG] Application ID: %s", appID)
	var params scalingo.AutoscalerUpdateParams
	changed := false

	if d.HasChange("metric") {
		params.Metric = stringAddr(d.Get("metric").(string))
		changed = true
	}

	if d.HasChange("target") {
		params.Target = float64Addr(d.Get("target").(float64))
		changed = true
	}

	if d.HasChange("min_containers") {
		params.MinContainers = intAddr(d.Get("min_containers").(int))
		changed = true
	}

	if d.HasChange("max_containers") {
		params.MaxContainers = intAddr(d.Get("max_containers").(int))
		changed = true
	}

	if d.HasChange("disabled") {
		params.Disabled = boolAddr(d.Get("disabled").(bool))
		changed = true
	}

	if changed {
		_, err := client.AutoscalerUpdate(appID, id, params)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[DEBUG] Autoscaler updated")

	return nil
}

func resourceAutoscalerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	appID, _ := d.Get("app").(string)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)
	log.Printf("[DEBUG] Application ID: %s", appID)

	client, _ := meta.(*scalingo.Client)
	err := client.AutoscalerRemove(appID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceAutoscalerImport is called when importing a new container_type
// resource. The ID must be "appID:autoscalerID" such as
// "5a155aa8f112e20010779b7a:sc-ac161a3d-d78b-4017-949b-19efb1e54083".
//
// Usage:
// $ terraform import terraform ID appID:autoscalerID
//
// Example:
// $ terraform import module.api.scalingo_autoscaler.api_autoweb 5a155aa8f112e20010779b7a:sc-ac161a3d-d78b-4017-949b-19efb1e54083
func resourceAutoscalerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("ID should have the following format: <app ID>:<autoscaler ID>")
	}
	appID := ids[0]
	id := ids[1]
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)

	d.SetId(id)
	err := d.Set("app", appID)
	if err != nil {
		return nil, err
	}

	diags := resourceAutoscalerRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
