package scalingo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func resourceScalingoAutoscaler() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAutoscalerCreate,
		ReadContext:   resourceAutoscalerRead,
		UpdateContext: resourceAutoscalerUpdate,
		DeleteContext: resourceAutoscalerDelete,
		Description:   "Resource representing an autoscaler of an application, setting rules to automatically scale up and down containers of the app",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID or Name of the targeted application",
			},
			"container_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Container type targeted by the autoscaler (web, worker, etc.)",
			},
			"min_containers": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Minimum number of containers (autoscaler won't get under it)",
			},
			"max_containers": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Maximum number of containers (autoscaler won't get over it)",
			},
			"metric": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Watched metric to base the autoscaling on (cpu, ram, etc.)",
			},
			"target": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Target reference value to base the autoscaling algorithm on",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable without deleting the autoscaler",
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

	autoscaler, err := client.AutoscalerAdd(ctx, appID, scalingo.AutoscalerAddParams{
		ContainerType: d.Get("container_type").(string),
		Metric:        d.Get("metric").(string),
		Target:        d.Get("target").(float64),
		MinContainers: d.Get("min_containers").(int),
		MaxContainers: d.Get("max_containers").(int),
	})
	if err != nil {
		return diag.Errorf("fail to get autoscaler information: %v", err)
	}

	d.SetId(autoscaler.ID)

	return nil
}

func resourceAutoscalerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	appID, _ := d.Get("app").(string)

	autoscaler, err := client.AutoscalerShow(ctx, appID, id)
	if err != nil {
		return diag.Errorf("fail to get autoscaler: %v", err)
	}

	err = SetAll(d, map[string]interface{}{
		"container_type": autoscaler.ContainerType,
		"min_containers": autoscaler.MinContainers,
		"max_containers": autoscaler.MaxContainers,
		"metric":         autoscaler.Metric,
		"target":         autoscaler.Target,
		"disabled":       autoscaler.Disabled,
	})
	if err != nil {
		return diag.Errorf("fail to store autoscaler informations: %v", err)
	}

	return nil
}

func resourceAutoscalerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	appID, _ := d.Get("app").(string)
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
		_, err := client.AutoscalerUpdate(ctx, appID, id, params)
		if err != nil {
			return diag.Errorf("fail to update autoscaler: %v", err)
		}
	}

	return nil
}

func resourceAutoscalerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	appID, _ := d.Get("app").(string)

	client, _ := meta.(*scalingo.Client)
	err := client.AutoscalerRemove(ctx, appID, id)
	if err != nil {
		return diag.Errorf("fail to destroy autoscaler: %v", err)
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
		return nil, fmt.Errorf("ID should have the following format: <app ID>:<autoscaler ID>")
	}
	appID := ids[0]
	id := ids[1]

	d.SetId(id)
	err := d.Set("app", appID)
	if err != nil {
		return nil, fmt.Errorf("fail to store app id: %v", err)
	}

	diags := resourceAutoscalerRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to read autoscaler informations: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
