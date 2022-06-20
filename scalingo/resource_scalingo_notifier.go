package scalingo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v4"
)

func resourceScalingoNotifier() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceScNotifierRead,
		CreateContext: resourceScNotifierCreate,
		UpdateContext: resourceScNotifierUpdate,
		DeleteContext: resourceScNotifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceScNotifierImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"app": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"send_all_alerts": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"send_all_events": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"selected_events": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"emails": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"webhook_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// resourceScNotifierRead performs the Scalingo API lookup
func resourceScNotifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	app, _ := d.Get("app").(string)
	notifier, err := client.NotifierByID(app, d.Id())
	if err != nil {
		return diag.Errorf("fail to find notifier %v of app %v: %v", app, d.Id(), err)
	}
	err = setFromScNotifier(d, client, notifier)
	if err != nil {
		return diag.Errorf("fail to set resource from API data: %v", err)
	}
	return nil
}

// resourceScNotifierCreate creates a notifier calling the Scalingo API
func resourceScNotifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	params, err := readNotifierParamsFromResource(d, client)
	if err != nil {
		return diag.Errorf("fail to read notifier params from resource: %v", err)
	}
	notifier, err := client.NotifierProvision(d.Get("app").(string), params)
	if err != nil {
		return diag.Errorf("fail to provision notifier: %v", err)
	}
	err = setFromScNotifier(d, client, notifier)
	if err != nil {
		return diag.Errorf("fail to set resource from API data: %v", err)
	}
	return nil
}

// resourceScNotifierUpdate updates a notifier calling the Scalingo API
func resourceScNotifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	params, err := readNotifierParamsFromResource(d, client)
	if err != nil {
		return diag.Errorf("fail to read notifier params from resource: %v", err)
	}
	notifier, err := client.NotifierUpdate(d.Get("app").(string), d.Id(), params)
	if err != nil {
		return diag.Errorf("fail to update notifier: %v", err)
	}
	err = setFromScNotifier(d, client, notifier)
	if err != nil {
		return diag.Errorf("fail to set resource from API data: %v", err)
	}
	return nil
}

func resourceScNotifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	err := client.NotifierDestroy(d.Get("app").(string), d.Id())
	if err != nil {
		return diag.Errorf("fail to delete notifier: %v", err)
	}
	return nil
}

func resourceScNotifierImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if !strings.Contains(d.Id(), ":") {
		return nil, errors.New("schema must be app_id:notifier_id")
	}
	split := strings.Split(d.Id(), ":")
	d.SetId(split[1])
	err := d.Set("app", split[0])
	if err != nil {
		return nil, fmt.Errorf("fail to set app id: %v", err)
	}

	diags := resourceScNotifierRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to import resource: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func readNotifierParamsFromResource(d *schema.ResourceData, c scalingo.API) (scalingo.NotifierParams, error) {
	types, err := c.EventTypesList()
	if err != nil {
		return scalingo.NotifierParams{}, fmt.Errorf("fail to list event types: %v", err)
	}

	sendAllEvents, _ := d.Get("send_all_events").(bool)
	sendAllAlerts, _ := d.Get("send_all_alerts").(bool)
	active, _ := d.Get("active").(bool)

	var selectedEventIDs []string
	selectedEvents, _ := d.Get("selected_events").(*schema.Set)
	for _, e := range selectedEvents.List() {
		for _, t := range types {
			if t.Name == e.(string) {
				selectedEventIDs = append(selectedEventIDs, t.ID)
				break
			}
		}
	}

	var userIDs []string
	for _, id := range d.Get("user_ids").([]interface{}) {
		userIDs = append(userIDs, id.(string))
	}

	var emails []string
	for _, email := range d.Get("emails").([]interface{}) {
		emails = append(emails, email.(string))
	}

	return scalingo.NotifierParams{
		PlatformID:       d.Get("platform_id").(string),
		Name:             d.Get("name").(string),
		SelectedEventIDs: selectedEventIDs,
		Active:           &active,
		SendAllEvents:    &sendAllEvents,
		SendAllAlerts:    &sendAllAlerts,

		// For email notifiers (email or user_id of collaborator)
		Emails:  emails,
		UserIDs: userIDs,

		// For Slack/Webhook/Rocket.Chat
		WebhookURL: d.Get("webhook_url").(string),
	}, nil
}

func setFromScNotifier(d *schema.ResourceData, c scalingo.API, notifier *scalingo.Notifier) error {
	types, err := c.EventTypesList()
	if err != nil {
		return fmt.Errorf("fail to list event types: %v", err)
	}

	d.SetId(notifier.ID)
	values := make(map[string]interface{})
	values["app"] = notifier.AppID
	values["platform_id"] = notifier.PlatformID
	if notifier.SendAllAlerts != nil {
		values["send_all_alerts"] = *notifier.SendAllAlerts
	}
	if notifier.SendAllEvents != nil {
		values["send_all_events"] = *notifier.SendAllEvents
	}
	values["active"] = notifier.Active

	selectedEvents := []string{}
	for _, sid := range notifier.SelectedEventIDs {
		for _, t := range types {
			if t.ID == sid {
				selectedEvents = append(selectedEvents, t.Name)
				break
			}
		}
	}
	values["selected_events"] = selectedEvents
	values["type"] = notifier.Type

	typeData := notifier.Specialize().TypeDataMap()
	for key, value := range typeData {
		values[key] = value
	}

	err = SetAll(d, values)
	if err != nil {
		return fmt.Errorf("fail to store values from notifer: %v", err)
	}
	return nil
}
