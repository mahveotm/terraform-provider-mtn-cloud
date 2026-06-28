# A monthly budget for 2026 — costs must have 12 entries (one per month).
resource "mtncloud_budget" "monthly" {
  name     = "platform-2026"
  scope    = "account"
  interval = "month"
  year     = "2026"
  currency = "NGN"
  costs    = [1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000, 1000]
}

# A single annual budget — costs has 1 entry.
resource "mtncloud_budget" "annual" {
  name     = "annual-cap-2026"
  interval = "year"
  year     = "2026"
  costs    = [12000]
}
