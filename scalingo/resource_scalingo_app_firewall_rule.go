package scalingo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v11"
	scalingohttp "github.com/Scalingo/go-scalingo/v11/http"
)

func resourceScalingoAppFirewallRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppFirewallRuleCreate,
		ReadContext:   resourceAppFirewallRuleRead,
		DeleteContext: resourceAppFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAppFirewallRuleImport,
		},
		Description: "Resource representing an application firewall rule. Changing the application, CIDR, or label recreates the rule.",
		Schema: map[string]*schema.Schema{
			"app": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the targeted application", //nolint:goconst // Keep the description next to its schema field.
			},
			"cidr": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "IPv4 CIDR to allow. Use /32 for a single IPv4 address.",
			},
			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional label attached to the rule",
			},
		},
	}
}

func resourceAppFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, _ := meta.(scalingo.AppsService)
	appID, _ := d.Get("app").(string)
	cidr, _ := d.Get("cidr").(string)
	label, _ := d.Get("label").(string)

	rule, err := client.AppsFirewallRuleCreate(ctx, appID, scalingo.AppFirewallRuleParams{
		CIDR:  cidr,
		Label: label,
	})
	if err != nil {
		return diag.Errorf("create app firewall rule: %v", err)
	}

	d.SetId(rule.ID)
	err = SetAll(d, map[string]any{
		"cidr":  rule.CIDR,
		"label": rule.Label,
	})
	if err != nil {
		return diag.Errorf("store app firewall rule information: %v", err)
	}
	return nil
}

func resourceAppFirewallRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, _ := meta.(scalingo.AppsService)
	appID, _ := d.Get("app").(string)

	rule, err := client.AppsFirewallRuleShow(ctx, appID, d.Id())
	if err != nil {
		var requestErr *scalingohttp.RequestFailedError
		if errors.As(err, &requestErr) && requestErr.Code == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("read app firewall rule: %v", err)
	}

	err = SetAll(d, map[string]any{
		"cidr":  rule.CIDR,
		"label": rule.Label,
	})
	if err != nil {
		return diag.Errorf("store app firewall rule information: %v", err)
	}
	return nil
}

func resourceAppFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, _ := meta.(scalingo.AppsService)
	appID, _ := d.Get("app").(string)

	err := client.AppsFirewallRuleDelete(ctx, appID, d.Id())
	if err != nil {
		var requestErr *scalingohttp.RequestFailedError
		if !errors.As(err, &requestErr) || requestErr.Code != http.StatusNotFound {
			return diag.Errorf("delete app firewall rule: %v", err)
		}
	}
	return nil
}

func resourceAppFirewallRuleImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	ids := strings.Split(d.Id(), ":")
	if len(ids) != 2 || ids[0] == "" || ids[1] == "" {
		return nil, errors.New("ID should have the following format: <app ID>:<firewall rule ID>")
	}

	d.SetId(ids[1])
	err := d.Set("app", ids[0])
	if err != nil {
		return nil, fmt.Errorf("store app id: %v", err)
	}
	err = DiagnosticError(resourceAppFirewallRuleRead(ctx, d, meta))
	if err != nil {
		return nil, fmt.Errorf("read app firewall rule: %v", err)
	}
	if d.Id() == "" {
		return nil, fmt.Errorf("app firewall rule %s not found for app %s", ids[1], ids[0])
	}
	return []*schema.ResourceData{d}, nil
}
