# Shell script task (inline content), run on the provisioned resource.
resource "mtncloud_task" "deploy" {
  name           = "deploy-app"
  type           = "shell"
  source_type    = "local"
  content        = file("${path.module}/deploy.sh")
  execute_target = "resource"
  sudo           = true
  retryable      = true
  retry_count    = 3
}

# Ansible playbook task sourced from a git integration/repository.
resource "mtncloud_task" "configure" {
  name     = "configure-web"
  type     = "ansible"
  git_id   = 12
  git_ref  = "main"
  playbook = "site.yml"
  tags     = "web"
}

# Email notification task.
resource "mtncloud_task" "notify" {
  name          = "notify-ops"
  type          = "email"
  source_type   = "local"
  content       = "<p>Provisioning complete.</p>"
  email_address = "ops@example.com"
  subject       = "Deployment finished"
}

# Write-attributes task: set custom attributes on the target during a workflow.
# `attributes` is a JSON object string; if the WAF rejects the JSON body, have it
# allow the payload.
resource "mtncloud_task" "tag" {
  name       = "tag-environment"
  type       = "write_attributes"
  attributes = jsonencode({ environment = "production" })
}

# Nested-workflow task: run an existing operational workflow as a task.
data "mtncloud_workflow" "ops" {
  name = "operational-maintenance"
}

resource "mtncloud_task" "run_ops" {
  name                    = "run-ops-workflow"
  type                    = "nested_workflow"
  operational_workflow_id = data.mtncloud_workflow.ops.id
}
