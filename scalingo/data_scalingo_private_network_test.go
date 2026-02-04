package scalingo

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccPrivateNetworkAppID(t *testing.T) string {
	appID := os.Getenv("SCALINGO_PRIVATE_NETWORK_APP_ID")
	if appID == "" {
		t.Skipf("SCALINGO_PRIVATE_NETWORK_APP_ID not set; skipping private network domain acceptance test")
	}
	return appID
}

func TestAccDataSourcePrivateNetworkDomain_basic(t *testing.T) {
	appID := testAccPrivateNetworkAppID(t)
	resourceName := "data.scalingo_private_network_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateNetworkDomainConfig(appID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app", appID),
					resource.TestCheckResourceAttrSet(resourceName, "domains.#"),
				),
			},
		},
	})
}

func testAccPrivateNetworkDomainConfig(appID string) string {
	return fmt.Sprintf(`
data "scalingo_private_network_domain" "test" {
  app = "%s"
}
`, appID)
}
