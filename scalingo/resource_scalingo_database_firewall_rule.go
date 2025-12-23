package scalingo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

const firewallRuleReadTimeout = 120 * time.Second

// firewallRuleWaitOpts returns wait options for firewall rule operations.
func firewallRuleWaitOpts(ruleID string) waitOptions {
	return waitOptions{
		timeout:    firewallRuleReadTimeout,
		timeoutErr: fmt.Errorf("firewall rule %s not visible before timeout", ruleID),
	}
}

func resourceScalingoDatabaseFirewallRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseFirewallRuleCreate,
		ReadContext:   resourceDatabaseFirewallRuleRead,
		DeleteContext: resourceDatabaseFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseFirewallRuleImport,
		},
		Description: "Resource representing a database firewall rule",

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Database NG",
			},
			"cidr": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"cidr", "managed_range_id"},
				ConflictsWith: []string{"managed_range_id"},
				Description:   "CIDR for a custom range firewall rule",
			},
			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Label for a custom range firewall rule",
			},
			"managed_range_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  []string{"cidr", "managed_range_id"},
				ConflictsWith: []string{"cidr"},
				Description:   "ID of a managed range firewall rule",
			},
			"rule_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the firewall rule (custom_range or managed_range)",
			},
		},
	}
}

func resourceDatabaseFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	databaseID, _ := d.Get("database_id").(string)
	cidr, _ := d.Get("cidr").(string)
	label, _ := d.Get("label").(string)
	managedRangeID, _ := d.Get("managed_range_id").(string)

	if cidr == "" && managedRangeID == "" {
		return diag.Errorf("one of cidr or managed_range_id must be set")
	}
	if cidr != "" && managedRangeID != "" {
		return diag.Errorf("cidr and managed_range_id are mutually exclusive")
	}
	if managedRangeID != "" && label != "" {
		return diag.Errorf("label can only be set with cidr")
	}

	appID, addonID, err := getDBAPIContext(ctx, client, databaseID)
	if err != nil {
		return diag.Errorf("resolve database context: %v", err)
	}

	params := scalingo.FirewallRuleCreateParams{}
	if managedRangeID != "" {
		params.Type = scalingo.FirewallRuleTypeManagedRange
		params.RangeID = managedRangeID
	} else {
		params.Type = scalingo.FirewallRuleTypeCustomRange
		params.CIDR = cidr
		params.Label = label
	}

	rule, err := previewClient.FirewallRulesCreate(ctx, appID, addonID, params)
	if err != nil {
		return diag.Errorf("create firewall rule: %v", err)
	}

	d.SetId(rule.ID)

	return resourceDatabaseFirewallRuleRead(ctx, d, meta)
}

func resourceDatabaseFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	databaseID, _ := d.Get("database_id").(string)

	appID, addonID, err := getDBAPIContext(ctx, client, databaseID)
	if err != nil {
		return diag.Errorf("resolve database context: %v", err)
	}

	var selected *scalingo.FirewallRule
	if d.IsNewResource() {
		var err error
		selected, err = waitUntilFirewallRuleVisible(ctx, previewClient, appID, addonID, d.Id())
		if err != nil {
			return diag.Errorf("wait for firewall rule: %v", err)
		}
	} else {
		var err error
		selected, err = findFirewallRule(ctx, previewClient, appID, addonID, d.Id())
		if err != nil {
			return diag.Errorf("list firewall rules: %v", err)
		}
	}

	if selected == nil {
		d.SetId("")
		return nil
	}

	err = SetAll(d, map[string]interface{}{
		"database_id":      databaseID,
		"cidr":             selected.CIDR,
		"label":            selected.Label,
		"managed_range_id": selected.RangeID,
		"rule_type":        string(selected.Type),
	})
	if err != nil {
		return diag.Errorf("store firewall rule information: %v", err)
	}

	return nil
}

func resourceDatabaseFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _ := meta.(*scalingo.Client)
	previewClient := scalingo.NewPreviewClient(client)

	databaseID, _ := d.Get("database_id").(string)

	appID, addonID, err := getDBAPIContext(ctx, client, databaseID)
	if err != nil {
		return diag.Errorf("resolve database context: %v", err)
	}

	err = previewClient.FirewallRulesDestroy(ctx, appID, addonID, d.Id())
	if err != nil {
		return diag.Errorf("destroy firewall rule: %v", err)
	}

	return nil
}

func resourceDatabaseFirewallRuleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 {
		return nil, errors.New("ID should have the following format: <database ID>:<firewall rule ID>")
	}

	databaseID := ids[0]
	ruleID := ids[1]

	d.SetId(ruleID)
	err := d.Set("database_id", databaseID)
	if err != nil {
		return nil, fmt.Errorf("fail to store database id: %v", err)
	}

	diags := resourceDatabaseFirewallRuleRead(ctx, d, meta)
	err = DiagnosticError(diags)
	if err != nil {
		return nil, fmt.Errorf("fail to read firewall rule: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func findFirewallRule(ctx context.Context, previewClient *scalingo.PreviewClient, appID, addonID, ruleID string) (*scalingo.FirewallRule, error) {
	rules, err := previewClient.FirewallRulesList(ctx, appID, addonID)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if rule.ID == ruleID {
			return &rule, nil
		}
	}

	return nil, nil
}

func waitUntilFirewallRuleVisible(ctx context.Context, previewClient *scalingo.PreviewClient, appID, addonID, ruleID string) (*scalingo.FirewallRule, error) {
	var result *scalingo.FirewallRule

	err := waitUntil(ctx, firewallRuleWaitOpts(ruleID), func() (bool, error) {
		rule, err := findFirewallRule(ctx, previewClient, appID, addonID, ruleID)
		if err != nil {
			return false, err
		}
		if rule != nil {
			result = rule
			return true, nil
		}
		return false, nil
	})

	return result, err
}
