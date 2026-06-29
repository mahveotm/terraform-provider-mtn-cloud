# A user group with a shared Linux login and sudo policy.
resource "mtncloud_user_group" "platform" {
  name         = "platform"
  description  = "Platform engineers"
  sudo_access  = true
  server_group = "platform"
  user_ids     = [mtncloud_user.jdoe.id]
}
