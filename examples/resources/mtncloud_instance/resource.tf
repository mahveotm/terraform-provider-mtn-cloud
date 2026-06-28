# Minimal instance: group, resource_pool, and security_group ("default") are
# inherited from the provider / defaulted. labels/tags merge with provider defaults.
resource "mtncloud_instance" "web" {
  name = "web-01"
  type = "MTN-CS10"
  plan = "G2S4"

  labels = ["web"]          # effective labels_all = ["terraform", "web"]
  tags   = { role = "web" } # tags_all merges provider default_tags
}
