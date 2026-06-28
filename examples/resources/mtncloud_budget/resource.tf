# A monthly budget for 2026 — costs must have 12 entries (one per month).
# `currency` is not set here: the MTN Cloud budget API ignores any requested
# currency and reports the account currency, so it is a read-only (computed)
# attribute you can reference but not configure.
resource "mtncloud_budget" "monthly" {
  name     = "platform-2026"
  scope    = "account"
  interval = "month"
  year     = "2026"
  costs    = [1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000]
}

# A single annual budget — costs has 1 entry.
resource "mtncloud_budget" "annual" {
  name     = "annual-cap-2026"
  interval = "year"
  year     = "2026"
  costs    = [12000]
}

# The account currency MTN Cloud applies to the budgets above (read-only).
output "budget_currency" {
  value = mtncloud_budget.annual.currency
}
