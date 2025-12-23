package scalingo

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Scalingo/go-scalingo/v8"
)

func intAddr(i int) *int {
	return &i
}

func uintAddr(i uint) *uint {
	return &i
}

func stringAddr(i string) *string {
	return &i
}

func boolAddr(i bool) *bool {
	return &i
}

func float64Addr(i float64) *float64 {
	return &i
}

func SetAll(d *schema.ResourceData, values map[string]interface{}) error {
	for name, value := range values {
		err := d.Set(name, value)
		if err != nil {
			return fmt.Errorf("fail to set field %s: %v", name, err)
		}
	}
	return nil
}

func DiagnosticError(diagnostics diag.Diagnostics) error {
	if len(diagnostics) == 0 {
		return nil
	}

	for _, d := range diagnostics {
		if d.Severity == diag.Error {
			return fmt.Errorf("%s %s", d.Summary, d.Detail)
		}
	}
	return nil
}

func getDBAPIContext(ctx context.Context, client *scalingo.Client, databaseID string) (string, string, error) {
	previewClient := scalingo.NewPreviewClient(client)

	database, err := previewClient.DatabaseShow(ctx, databaseID)
	if err != nil {
		return "", "", fmt.Errorf("get database information for %v: %v", databaseID, err)
	}

	appID := database.App.ID
	if appID == "" {
		return "", "", fmt.Errorf("no app id found for database %v", databaseID)
	}

	addons, err := client.AddonsList(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("list addons: %v", err)
	}
	if len(addons) == 0 {
		return "", "", fmt.Errorf("no addons found for database application %v", appID)
	}

	return appID, addons[0].ID, nil
}
