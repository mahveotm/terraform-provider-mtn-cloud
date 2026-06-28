resource "mtncloud_scale_threshold" "web_cpu" {
  name           = "web-cpu-autoscale"
  auto_upscale   = true
  auto_downscale = true
  min_count      = 1
  max_count      = 5

  enable_cpu_threshold = true
  min_cpu_percentage   = 30
  max_cpu_percentage   = 75
}
