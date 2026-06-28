resource "mtncloud_wiki_page" "runbook" {
  name     = "Deployment Runbook"
  category = "ops"
  content  = <<-EOT
    # Deployment Runbook

    1. Provision the instance.
    2. Attach the network.
    3. Apply security groups.
  EOT
}
