# A user role with a permission set in the Morpheus role API shape.
resource "mtncloud_role" "operators" {
  name        = "operators"
  description = "Day-to-day operators"

  permission_set = jsonencode({
    globalSiteAccess         = "all"
    globalInstanceTypeAccess = "full"
    featurePermissions = [
      { code = "admin-users", access = "full" },
      { code = "infrastructure-network", access = "read" },
    ]
  })
}
