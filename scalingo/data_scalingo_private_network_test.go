package scalingo

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccPrivateNetworkAppID(t *testing.T) string {
	appID := os.Getenv("SCALINGO_PRIVATE_NETWORK_APP_ID")
	if appID == "" {
		t.Skipf("SCALINGO_PRIVATE_NETWORK_APP_ID not set; skipping private network domain acceptance test")
	}
	return appID
}

func TestAccDataSourcePrivateNetworkDomains(t *testing.T) {
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
					// ".#" is a Terraform convention for counting elements in a set or list attribute
					resource.TestCheckResourceAttrSet(resourceName, "domains.#"),
				),
			},
			{
				Config: testAccPrivateNetworkDomainPaginationConfig(appID, 1, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app", appID),
					testAccCheckDomainsCountMax(resourceName, 1),
				),
			},
			{
				Config: testAccPrivateNetworkDomainPaginationConfig(appID, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app", appID),
					testAccCheckDomainsCountMax(resourceName, 1),
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

func testAccPrivateNetworkDomainPaginationConfig(appID string, page int, pageSize int) string {
	return fmt.Sprintf(`
data "scalingo_private_network_domain" "test" {
  app       = "%s"
  page      = %d
  page_size = %d
}
`, appID, page, pageSize)
}

func testAccCheckDomainsCountMax(resourceName string, max int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.New("resource not found: " + resourceName)
		}
		countStr, ok := rs.Primary.Attributes["domains.#"]
		if !ok || countStr == "" {
			return errors.New("domains.# not set")
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("invalid domains.# value %q: %v", countStr, err)
		}
		if count > max {
			return fmt.Errorf("expected domains count <= %d, got %d", max, count)
		}
		return nil
	}
}
