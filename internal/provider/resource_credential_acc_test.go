package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccCredentialResource is the reference acceptance test every CRUD resource
// follows: create + assert, update in place, and import round-trip (ignoring
// write-only secret fields the API never returns). Runs only under TF_ACC with a
// valid MTN_CLOUD_TOKEN; the sweeper "mtncloud_credential" cleans up leftovers.
func TestAccCredentialResource(t *testing.T) {
	name := accName("cred")
	const addr = "mtncloud_credential.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // create
				Config: testAccCredentialConfig(name, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", name),
					resource.TestCheckResourceAttr(addr, "type", "username-password"),
					resource.TestCheckResourceAttr(addr, "description", "first"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{ // update description in place (no replacement)
				Config: testAccCredentialConfig(name, "second"),
				Check:  resource.TestCheckResourceAttr(addr, "description", "second"),
			},
			{ // import round-trip; password is write-only and never returned
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccCredentialConfig(name, description string) string {
	return testAccProviderConfig + fmt.Sprintf(`
resource "mtncloud_credential" "test" {
  type        = "username-password"
  name        = %q
  description = %q
  username    = "acc-user"
  password    = "acc-pass"
}
`, name, description)
}
