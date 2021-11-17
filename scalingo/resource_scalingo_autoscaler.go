package scalingo

import (
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo"
)

func resourceScalingoAutoscaler() *schema.Resource {
	return &schema.Resource{
		Create: resourceAutoscalerCreate,
		Read:   resourceAutoscalerRead,
		Update: resourceAutoscalerUpdate,
		Delete: resourceAutoscalerDelete,

		Schema: map[string]*schema.Schema{
			"app": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"min_containers": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"max_containers": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"metric": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"target": &schema.Schema{
				Type:     schema.TypeFloat,
				Required: true,
			},
			"disabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			State: resourceAutoscalerImport,
		},
	}
}

func resourceAutoscalerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	appID := d.Get("app").(string)
	log.Printf("[DEBUG] Application ID: %s", appID)

	autoscaler, err := client.AutoscalerAdd(appID, scalingo.AutoscalerAddParams{
		ContainerType: d.Get("container_type").(string),
		Metric:        d.Get("metric").(string),
		Target:        d.Get("target").(float64),
		MinContainers: d.Get("min_containers").(int),
		MaxContainers: d.Get("max_containers").(int),
	})
	if err != nil {
		return err
	}

	d.SetId(autoscaler.ID)

	return nil
}

func resourceAutoscalerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	id := d.Id()
	appID := d.Get("app").(string)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)
	log.Printf("[DEBUG] Application ID: %s", appID)

	autoscaler, err := client.AutoscalerShow(appID, id)
	if err != nil {
		log.Printf("[INFO] Error getting autoscaler %#v", err)
		return err
	}
	log.Printf("[INFO] Successfuly fetched autoscaler")

	d.Set("container_type", autoscaler.ContainerType)
	d.Set("min_containers", autoscaler.MinContainers)
	d.Set("max_containers", autoscaler.MaxContainers)
	d.Set("metric", autoscaler.Metric)
	d.Set("target", autoscaler.Target)
	d.Set("disabled", autoscaler.Disabled)

	return nil
}

func resourceAutoscalerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*scalingo.Client)

	id := d.Id()
	appID := d.Get("app").(string)
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
			return err
		}
	}

	log.Printf("[DEBUG] Autoscaler updated")

	return nil
}

func resourceAutoscalerDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	appID := d.Get("app").(string)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)
	log.Printf("[DEBUG] Application ID: %s", appID)

	client := meta.(*scalingo.Client)
	err := client.AutoscalerRemove(appID, id)
	if err != nil {
		return err
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
func resourceAutoscalerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("ID should have the following format: <app ID>:<autoscaler ID>")
	}
	appID := ids[0]
	id := ids[1]
	log.Printf("[DEBUG] Application ID: %s", appID)
	log.Printf("[DEBUG] Autoscaler ID: %s", id)

	d.SetId(id)
	d.Set("app", appID)

	err := resourceAutoscalerRead(d, meta)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
