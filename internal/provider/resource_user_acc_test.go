package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// TestAccUserResource creates a role -> user, imports it, and asserts a username
// rename plans as an in-place Update. The rename is asserted plan-only for the same
// reason as the role test: MTN's user endpoint has a sub-second read-after-write lag
// that the harness's instant post-apply refresh can read stale (real users won't).
// Passwords are write-only, so they're ignored on verify.
func TestAccUserResource(t *testing.T) {
	name := accName("user")
	renamed := accName("user")
	const addr = "mtncloud_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccUserConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "username", name),
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestCheckResourceAttr(addr, "role_ids.#", "1"),
				),
			},
			{ // import: passwords are write-only and never returned
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "linux_password", "windows_password", "password_expired"},
			},
			{ // rename guard: a username change must plan as an in-place Update, never a replacement
				Config:             testAccUserConfig(renamed),
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

func testAccUserConfig(name string) string {
	return testAccProviderConfig + fmt.Sprintf(`
resource "mtncloud_role" "for_user" {
  name = "%s-role"
}

resource "mtncloud_user" "test" {
  username = %q
  email    = "%s@example.invalid"
  password = "Tf!acc-Passw0rd"
  role_ids = [mtncloud_role.for_user.id]
}
`, name, name, name)
}
