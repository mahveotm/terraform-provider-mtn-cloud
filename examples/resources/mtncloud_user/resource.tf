# A user assigned to one or more roles. The password is write-only.
resource "mtncloud_role" "operators" {
  name = "operators"
}

resource "mtncloud_user" "jdoe" {
  username = "jdoe"
  email    = "jdoe@example.com"
  password = var.user_password # set via a variable / secret, not in plain config

  first_name = "Jane"
  last_name  = "Doe"
  role_ids   = [mtncloud_role.operators.id]
}

variable "user_password" {
  type      = string
  sensitive = true
}
