package scalingo

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"scalingo": func() (*schema.Provider, error) { //nolint:unparam // required by terraform-plugin-sdk
		return Provider(), nil
	},
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance tests")
	}
	if os.Getenv("SCALINGO_API_TOKEN") == "" {
		t.Fatalf("SCALINGO_API_TOKEN must be set for acceptance tests")
	}
}
