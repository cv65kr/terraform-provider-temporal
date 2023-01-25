package temporal

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestTemporal_Resource_Namespace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "temporal_namespace" "ns_test1" {
						name = "test1"
						description = "Test namespace 1"
						owner_email = "test@example.com"
						retention = "240"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "id", "test1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "name", "test1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "description", "Test namespace 1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "owner_email", "test@example.com"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "retention", "240"),
				),
			},
		},
	})
}
