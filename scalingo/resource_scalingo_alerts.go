package scalingo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	scalingo "github.com/Scalingo/go-scalingo/v5"
)

func resourceScalingoAlerts() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertsCreate,
		ReadContext:   resourceAlertsRead,
		UpdateContext: resourceAlertsUpdate,
		DeleteContext: resourceAlertsDelete,

		Schema: map[string]*schema.Schema{
			"app": {
				Type:     schema.TypeString,
				Required: true,
			},
			"container_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"metric": {
				Type:     schema.TypeString,
				Required: true,
			},
			"limit": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"send_when_below": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"duration_before_trigger": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"remind_every": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notifiers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceAlertsImport,
		},
	}
}

func resourceAlertsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	app, _ := d.Get("app").(string)

	remindEvery, err := time.ParseDuration(d.Get("remind_every").(string))
	if err != nil {
		return diag.Errorf("fail to parse remind_every information: %v", err)
	}
	durationBeforeTrigger, err := time.ParseDuration(d.Get("duration_before_trigger").(string))
	if err != nil {
		return diag.Errorf("fail to parse duration_before_trigger information: %v", err)
	}
	alert, err := client.AlertAdd(ctx, app, scalingo.AlertAddParams{
		ContainerType:         d.Get("container_type").(string),
		Metric:                d.Get("metric").(string),
		Limit:                 d.Get("limit").(float64),
		Disabled:              d.Get("disabled").(bool),
		RemindEvery:           &remindEvery,
		DurationBeforeTrigger: &durationBeforeTrigger,
		SendWhenBelow:         d.Get("send_when_below").(bool),
		Notifiers:             d.Get("notifiers").([]string),
	})
	if err != nil {
		return diag.Errorf("fail to create alert: %v", err)
	}
	d.SetId(alert.ID)

	return nil
}

func resourceAlertsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)

	alert, err := client.AlertShow(ctx, app, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(alert.ID)
	err = SetAll(d, map[string]interface{}{
		"app":                     alert.AppID,
		"container_type":          alert.ContainerType,
		"metric":                  alert.Metric,
		"limit":                   alert.Limit,
		"disabled":                alert.Disabled,
		"send_when_below":         alert.SendWhenBelow,
		"duration_before_trigger": alert.DurationBeforeTrigger.String(),
		"remind_every":            alert.RemindEvery,
		"notifiers":               alert.Notifiers,
		"created_at":              alert.CreatedAt,
		"updated_at":              alert.UpdatedAt,
		"metadata":                alert.Metadata,
	})
	if err != nil {
		return diag.Errorf("fail to get alerts information: %v", err)
	}

	return nil
}

func resourceAlertsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	id := d.Id()
	app, _ := d.Get("app").(string)

	alertUpdateParams := scalingo.AlertUpdateParams{}
	changed := false
	if d.HasChange("container_type") {
		alertUpdateParams.ContainerType = stringAddr(d.Get("container_type").(string))
		changed = true
	}
	if d.HasChange("metric") {
		alertUpdateParams.Metric = stringAddr(d.Get("metric").(string))
		changed = true
	}
	if d.HasChange("limit") {
		alertUpdateParams.Limit = float64Addr(d.Get("limit").(float64))
		changed = true
	}
	if d.HasChange("disabled") {
		alertUpdateParams.Disabled = boolAddr(d.Get("disabled").(bool))
		changed = true
	}
	durationBeforeTrigger, _ := time.ParseDuration(d.Get("duration_before_trigger").(string))
	if d.HasChange("duration_before_trigger") {
		alertUpdateParams.DurationBeforeTrigger = &durationBeforeTrigger
		changed = true
	}
	remindEvery, _ := time.ParseDuration(d.Get("remind_every").(string))
	if d.HasChange("remind_every") {
		alertUpdateParams.RemindEvery = &remindEvery
		changed = true
	}
	if d.HasChange("send_when_below") {
		alertUpdateParams.SendWhenBelow = boolAddr(d.Get("send_when_below").(bool))
		changed = true
	}
	notifiers, _ := d.Get("send_when_below").([]string)
	if d.HasChange("notifiers") {
		alertUpdateParams.Notifiers = &notifiers
		changed = true
	}

	if changed {
		alertUpdate, err := client.AlertUpdate(ctx, app, id, alertUpdateParams)
		if err != nil {
			return diag.Errorf("fail to update alerts: %v", err)
		}
		d.SetId(alertUpdate.ID)
		err = SetAll(d, map[string]interface{}{
			"app":                     alertUpdate.AppID,
			"container_type":          alertUpdate.ContainerType,
			"metric":                  alertUpdate.Metric,
			"limit":                   alertUpdate.Limit,
			"disabled":                alertUpdate.Disabled,
			"send_when_below":         alertUpdate.SendWhenBelow,
			"duration_before_trigger": alertUpdate.DurationBeforeTrigger.String(),
			"remind_every":            alertUpdate.RemindEvery,
			"notifiers":               alertUpdate.Notifiers,
			"created_at":              alertUpdate.CreatedAt,
			"updated_at":              alertUpdate.UpdatedAt,
			"metadata":                alertUpdate.Metadata,
		})
		if err != nil {
			return diag.Errorf("fail to get alerts information: %v", err)
		}
	}

	return nil
}

func resourceAlertsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)

	err := client.AlertRemove(ctx, app, d.Id())
	if err != nil {
		return diag.Errorf("fail to delete alert: %v", err)
	}
	return nil
}

func resourceAlertsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if !strings.Contains(d.Id(), ":") {
		return nil, fmt.Errorf("schema must be app_id:alert_id")
	}
	split := strings.Split(d.Id(), ":")
	d.SetId(split[1])
	err := d.Set("app", split[0])
	if err != nil {
		return nil, fmt.Errorf("fail to set app id: %v", err)
	}

	diags := resourceAlertsRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to import resource: %v", err)
	}

	return []*schema.ResourceData{d}, nil

}
