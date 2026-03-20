package scalingo

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderClientConfigSetsUserAgent(t *testing.T) {
	currentVersion := Version
	Version = "test-version"
	t.Cleanup(func() {
		Version = currentVersion
	})

	data := schema.TestResourceDataRaw(t, Provider().Schema, map[string]any{
		"region": "osc-fr1",
	})

	config := providerClientConfig(data)

	if config.UserAgent != "Scalingo Terraform Provider vtest-version" {
		t.Fatalf("unexpected user agent: %q", config.UserAgent)
	}
}
