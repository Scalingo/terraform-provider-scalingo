package scalingo

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			return err
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
