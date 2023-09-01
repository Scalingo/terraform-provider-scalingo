package scalingo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v6"
)

func resourceScalingoLogDrain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLogDrainCreate,
		ReadContext:   resourceLogDrainRead,
		DeleteContext: resourceLogDrainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceLogDrainImporter,
		},
		Description: "Resource representing a log drain attached to an application or a database",
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "Type of log drain (syslog/http/elk/etc.)",
			},
			"host": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "hostname to which logs are transferred",
			},
			"port": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "port to which logs are transferred",
			},
			"token": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Description: "token used to authenticate log drain requests (see compatible destination in documentation)",
			},
			"drain_region": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "destination region (see compatible destination in documentation)",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "helper to generate all the host/port/token fields",
			},
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application",
			},
			"addon": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "UUID of the targeted addon attached to the application",
			},
			"drain_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "URL representation of the log drain",
			},
		},
	}
}

func resourceLogDrainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)

	logDrainType, _ := d.Get("type").(string)
	host, _ := d.Get("host").(string)
	port, _ := d.Get("port").(string)
	token, _ := d.Get("token").(string)
	drainRegion, _ := d.Get("drain_region").(string)
	url, _ := d.Get("url").(string)

	params := scalingo.LogDrainAddParams{
		Type:        logDrainType,
		Host:        host,
		Port:        port,
		Token:       token,
		DrainRegion: drainRegion,
		URL:         url,
	}

	var res *scalingo.LogDrainRes
	var err error

	addonID, ok := d.Get("addon").(string)
	if ok && addonID != "" {
		res, err = client.LogDrainAddonAdd(ctx, appID, addonID, params)
	} else {
		res, err = client.LogDrainAdd(ctx, appID, params)
	}
	if err != nil {
		return diag.Errorf("fail to create log drain: %v", err)
	}

	err = d.Set("drain_url", res.Drain.URL)
	if err != nil {
		return diag.Errorf("fail to set drain_url: %v", err)
	}

	d.SetId(computeID(d))

	return nil
}

func resourceLogDrainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	drainURL, _ := d.Get("drain_url").(string)

	if drainURL == "" {
		d.MarkNewResource()
		return nil
	}

	var err error
	var res []scalingo.LogDrain
	addonID, ok := d.Get("addon").(string)
	if ok && addonID != "" {
		res, err = client.LogDrainsAddonList(ctx, appID, addonID)
	} else {
		res, err = client.LogDrainsList(ctx, appID)
	}
	if err != nil {
		return diag.Errorf("fail to list log drains: %v", err)
	}

	var logDrain *scalingo.LogDrain
	for _, drain := range res {
		if drain.URL == drainURL {
			if logDrain != nil {
				return diag.Errorf("ambiguous response: to log drains has the same URL")
			}
			logDrain = &drain
		}
	}

	if logDrain == nil || drainURL != logDrain.URL {
		d.MarkNewResource()
	}
	return nil
}

func resourceLogDrainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)

	appID, _ := d.Get("app").(string)
	drainURL, _ := d.Get("drain_url").(string)

	if drainURL == "" {
		return diag.Errorf("no drain_url set")
	}

	var err error
	addonID, ok := d.Get("addon").(string)
	if ok && addonID != "" {
		err = client.LogDrainAddonRemove(ctx, appID, addonID, drainURL)
	} else {
		err = client.LogDrainRemove(ctx, appID, drainURL)
	}
	if err != nil {
		return diag.Errorf("fail to destroy log drain: %v", err)
	}

	return nil
}

func resourceLogDrainImporter(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if !strings.Contains(d.Id(), "#") {
		return nil, fmt.Errorf("schema must be app_id[#addon_id]#drain_url")
	}
	split := strings.Split(d.Id(), "#")
	if len(split) <= 1 || len(split) > 3 {
		return nil, fmt.Errorf("schema must be app_id[#addon_id]#drain_url")
	}

	values := make(map[string]interface{})
	values["app"] = split[0]
	values["drain_url"] = split[len(split)-1]
	if len("split") == 3 {
		values["addon"] = split[2]
	}

	err := SetAll(d, values)
	if err != nil {
		return nil, fmt.Errorf("fail to save values: %v", err)
	}

	d.SetId(computeID(d))

	return []*schema.ResourceData{d}, nil
}

func computeID(d *schema.ResourceData) string {
	appID, _ := d.Get("app").(string)
	addonID, _ := d.Get("addon").(string)
	drainURL, _ := d.Get("drain_url").(string)
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", appID, addonID, drainURL)))
	return hex.EncodeToString(hash[:])
}
