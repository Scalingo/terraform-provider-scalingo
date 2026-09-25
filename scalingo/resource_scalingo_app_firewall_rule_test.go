package scalingo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Scalingo/go-scalingo/v11"
	scalingohttp "github.com/Scalingo/go-scalingo/v11/http"
	"github.com/Scalingo/go-scalingo/v11/scalingomock"
)

func TestAppFirewallRuleCreate(t *testing.T) {
	const appID = "app-123"
	const ruleID = "rule-456"
	rule := scalingo.AppFirewallRule{ID: ruleID, AppID: appID, CIDR: "192.0.2.0/24", Label: "office"}
	ctx := t.Context()
	client := scalingomock.NewMockAppsService(gomock.NewController(t))
	client.EXPECT().AppsFirewallRuleCreate(ctx, appID, scalingo.AppFirewallRuleParams{
		CIDR: rule.CIDR, Label: rule.Label,
	}).Return(&rule, nil)
	resource := resourceScalingoAppFirewallRule()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"app": appID, "cidr": rule.CIDR, "label": rule.Label,
	})
	diags := resource.CreateContext(ctx, data, client)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, ruleID, data.Id())
	assert.Equal(t, rule.CIDR, data.Get("cidr"))
	assert.Equal(t, rule.Label, data.Get("label"))
}

func TestAppFirewallRuleReadUpdatedRule(t *testing.T) {
	const appID = "app-123"
	const ruleID = "rule-456"
	updatedRule := scalingo.AppFirewallRule{ID: ruleID, AppID: appID, CIDR: "198.51.100.4/32"}
	ctx := t.Context()
	client := scalingomock.NewMockAppsService(gomock.NewController(t))
	client.EXPECT().AppsFirewallRuleShow(ctx, appID, ruleID).Return(&updatedRule, nil)
	resource := resourceScalingoAppFirewallRule()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"app": appID, "cidr": "192.0.2.0/24", "label": "office",
	})
	data.SetId(ruleID)

	// Refresh must pick up remote changes, including removal of an optional label.
	diags := resource.ReadContext(ctx, data, client)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, ruleID, data.Id())
	assert.Equal(t, updatedRule.CIDR, data.Get("cidr"))
	assert.Empty(t, data.Get("label"))
}

func TestAppFirewallRuleImport(t *testing.T) {
	const appID = "app-123"
	const ruleID = "rule-456"
	rule := scalingo.AppFirewallRule{ID: ruleID, AppID: appID, CIDR: "198.51.100.4/32"}
	ctx := t.Context()
	client := scalingomock.NewMockAppsService(gomock.NewController(t))
	client.EXPECT().AppsFirewallRuleShow(ctx, appID, ruleID).Return(&rule, nil)
	resource := resourceScalingoAppFirewallRule()
	imported := schema.TestResourceDataRaw(t, resource.Schema, nil)
	imported.SetId(appID + ":" + ruleID)
	states, err := resource.Importer.StateContext(ctx, imported, client)
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.Equal(t, ruleID, states[0].Id())
	assert.Equal(t, appID, states[0].Get("app"))
	assert.Equal(t, rule.CIDR, states[0].Get("cidr"))
}

func TestAppFirewallRuleDelete(t *testing.T) {
	const appID = "app-123"
	const ruleID = "rule-456"
	ctx := t.Context()
	client := scalingomock.NewMockAppsService(gomock.NewController(t))
	client.EXPECT().AppsFirewallRuleDelete(ctx, appID, ruleID).Return(nil)
	resource := resourceScalingoAppFirewallRule()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"app": appID, "cidr": "192.0.2.0/24",
	})
	data.SetId(ruleID)

	diags := resource.DeleteContext(ctx, data, client)
	require.False(t, diags.HasError(), "%v", diags)
}

func TestAppFirewallRuleAPIErrors(t *testing.T) {
	for _, operation := range []struct {
		name string
		call func(context.Context, *schema.ResourceData, any) diag.Diagnostics
	}{
		{"create", resourceAppFirewallRuleCreate},
		{"read", resourceAppFirewallRuleRead},
		{"delete", resourceAppFirewallRuleDelete},
	} {
		for _, status := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
			t.Run(fmt.Sprintf("%s/%d", operation.name, status), func(t *testing.T) {
				ctx := t.Context()
				client := scalingomock.NewMockAppsService(gomock.NewController(t))
				apiErr := &scalingohttp.RequestFailedError{Code: status, APIError: errors.New("API failure")}
				switch operation.name {
				case "create":
					client.EXPECT().AppsFirewallRuleCreate(ctx, "app-123", scalingo.AppFirewallRuleParams{
						CIDR: "192.0.2.0/24",
					}).Return(nil, apiErr)
				case "read":
					client.EXPECT().AppsFirewallRuleShow(ctx, "app-123", "rule-456").Return(nil, apiErr)
				case "delete":
					client.EXPECT().AppsFirewallRuleDelete(ctx, "app-123", "rule-456").Return(apiErr)
				}
				data := schema.TestResourceDataRaw(t, resourceScalingoAppFirewallRule().Schema, map[string]any{
					"app": "app-123", "cidr": "192.0.2.0/24",
				})
				if operation.name != "create" {
					data.SetId("rule-456")
				}
				diags := operation.call(ctx, data, client)
				wantError := status != http.StatusNotFound || operation.name == "create"
				assert.Equal(t, wantError, diags.HasError(), "%v", diags)
				wantID := "rule-456"
				if operation.name == "create" || (operation.name == "read" && status == http.StatusNotFound) {
					wantID = ""
				}
				assert.Equal(t, wantID, data.Id())
			})
		}
	}
}

func TestAppFirewallRuleImportInvalidID(t *testing.T) {
	for _, id := range []string{"", "rule-456", ":rule-456", "app-123:", "app:rule:extra"} {
		t.Run(id, func(t *testing.T) {
			resource := resourceScalingoAppFirewallRule()
			data := schema.TestResourceDataRaw(t, resource.Schema, nil)
			data.SetId(id)
			_, err := resource.Importer.StateContext(t.Context(), data, nil)
			assert.Error(t, err, "expected invalid import ID to be rejected")
		})
	}
}
