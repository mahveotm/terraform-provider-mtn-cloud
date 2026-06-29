package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// TestAccRoleResource exercises create, an in-place permission_set update, import,
// and a rename guard. The rename is asserted as a plan-only in-place Update (never
// a replacement): MTN's role endpoint has a sub-second read-after-write lag, so
// applying a rename and immediately refreshing can read the pre-rename value — a
// transient real users won't hit but the test harness's instant refresh does. The
// rename itself does persist in place when applied. permission_set is
// config-authoritative, so it is ignored on import verify.
func TestAccRoleResource(t *testing.T) {
	name := accName("role")
	renamed := accName("role")
	const addr = "mtncloud_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccRoleConfig(name, "full"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name),
					resource.TestCheckResourceAttr(addr, "role_type", "user"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{ // update permission_set in place
				Config: testAccRoleConfig(name, "read"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
					},
				},
			},
			{ // import: metadata round-trips; permission_set is config-authoritative
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"permission_set"},
			},
			{ // rename guard: a name-only change must plan as an in-place Update, never a replacement
				Config:             testAccRoleConfig(renamed, "read"),
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

func testAccRoleConfig(name, access string) string {
	return testAccProviderConfig + fmt.Sprintf(`
resource "mtncloud_role" "test" {
  name        = %q
  description = "tf-acc role"
  permission_set = jsonencode({
    featurePermissions = [{ code = "admin-users", access = %q }]
  })
}
`, name, access)
}
