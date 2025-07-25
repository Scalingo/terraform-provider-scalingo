package scalingo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func resourceScalingoAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertsCreate,
		ReadContext:   resourceAlertsRead,
		UpdateContext: resourceAlertsUpdate,
		DeleteContext: resourceAlertsDelete,
		Description:   "Resource representing an alert definition for an application",

		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID or Name of the targeted application",
			},
			"container_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of containers (web/worker/...) watched by the alert",
			},
			"metric": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Metric (cpu/ram/swap/rpm/...) monitored by the alert",
			},
			"limit": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Limit/Threshold value at which the alert is triggered",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable the alert without deleting it",
			},
			"send_when_below": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Send alert when the current value is below the threshold instead of above",
			},
			"duration_before_trigger": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "0s",
				Description: "Delay before triggering the alert when the threshold is reached",
			},
			"remind_every": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Duration after which the alert will be re-triggered and sent to the notifiers",
			},
			"notifiers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of notifier IDs to which alerts are sent",
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

	var (
		remindEvery           time.Duration
		durationBeforeTrigger time.Duration
		err                   error
	)

	params := scalingo.AlertAddParams{
		ContainerType: d.Get("container_type").(string),
		Metric:        d.Get("metric").(string),
		Limit:         d.Get("limit").(float64),
		Disabled:      d.Get("disabled").(bool),
		SendWhenBelow: d.Get("send_when_below").(bool),
	}

	for _, notifierID := range d.Get("notifiers").([]interface{}) {
		params.Notifiers = append(params.Notifiers, notifierID.(string))
	}

	remindEveryStr, _ := d.Get("remind_every").(string)
	if remindEveryStr != "" {
		remindEvery, err = time.ParseDuration(remindEveryStr)
		if err != nil {
			return diag.Errorf("fail to parse remind_every information: %v", err)
		}
	}

	durationBeforeTriggerStr, _ := d.Get("duration_before_trigger").(string)
	if durationBeforeTriggerStr != "" {
		durationBeforeTrigger, err = time.ParseDuration(durationBeforeTriggerStr)
		if err != nil {
			return diag.Errorf("fail to parse duration_before_trigger information: %v", err)
		}
	}

	if remindEvery != 0 {
		params.RemindEvery = &remindEvery
	}
	if durationBeforeTrigger != 0 {
		params.DurationBeforeTrigger = &durationBeforeTrigger
	}

	alert, err := client.AlertAdd(ctx, app, params)
	if err != nil {
		return diag.Errorf("fail to create alert: %v", err)
	}
	d.SetId(alert.ID)

	return nil
}

func resourceAlertsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	app, _ := d.Get("app").(string)

	alerts, err := client.AlertsList(ctx, app)
	if err != nil {
		return diag.FromErr(err)
	}
	filteredAlerts := keepIf(alerts, func(a *scalingo.Alert) bool {
		return a.ID == d.Id()
	})
	if len(filteredAlerts) != 1 {
		return diag.Errorf("fail to get alerts information: %v", err)
	}
	alert := filteredAlerts[0]
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
	if d.HasChange("notifiers") {
		alertUpdateParams.Notifiers = new([]string)
		for _, notifierID := range d.Get("notifiers").([]interface{}) {
			*alertUpdateParams.Notifiers = append(*alertUpdateParams.Notifiers, notifierID.(string))
		}
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
			"notifiers":               alertUpdateParams.Notifiers,
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
