package temporal

import (
	"testing"
	"time"

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
						clusters = ["active"]
						owner_email = "test@example.com"
						retention = "17"

						namespace_data = {
							"k1": "v1",
							"k2": "v2"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "id", "test1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "name", "test1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "description", "Test namespace 1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "owner_email", "test@example.com"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "namespace_data.k1", "v1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "namespace_data.k2", "v2"),
				),
			},
			{
				PreConfig: func() {
					// Temporal need some time to create namespace
					time.Sleep(10 * time.Second)
				},
				Config: `
					resource "temporal_namespace" "ns_test1" {
						name = "test1"
						description = "Test namespace 1 with small change"
						owner_email = "test2@example.com"
						clusters = ["active"]
						retention = "10"

						namespace_data = {
							"k1": "v1",
							"k2": "v3",
							"k4": "v4"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "id", "test1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "name", "test1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "description", "Test namespace 1 with small change"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "owner_email", "test2@example.com"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "retention", "10"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "namespace_data.k1", "v1"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "namespace_data.k2", "v3"),
					resource.TestCheckResourceAttr("temporal_namespace.ns_test1", "namespace_data.k4", "v4"),
				),
			},
		},
	})
}
