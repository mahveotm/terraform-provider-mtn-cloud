package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// TestAccUserGroupResource covers create, import, and a rename/attribute change
// asserted as an in-place Update. The change is asserted plan-only for the same
// reason as the role test (MTN's read-after-write lag on an instant refresh).
func TestAccUserGroupResource(t *testing.T) {
	name := accName("ug")
	renamed := accName("ug")
	const addr = "mtncloud_user_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccUserGroupConfig(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name),
					resource.TestCheckResourceAttr(addr, "sudo_access", "false"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{ // import round-trip
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // rename + sudo change guard: must plan as an in-place Update, never a replacement
				Config:             testAccUserGroupConfig(renamed, true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccUserGroupConfig(name string, sudo bool) string {
	return testAccProviderConfig + fmt.Sprintf(`
resource "mtncloud_user_group" "test" {
  name        = %q
  description = "tf-acc user group"
  sudo_access = %t
}
`, name, sudo)
}
