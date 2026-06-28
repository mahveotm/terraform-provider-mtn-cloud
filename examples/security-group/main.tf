terraform {
  required_providers {
    mtncloud = {
      source  = "mahveotm/mtn-cloud"
      version = "0.1.0"
    }
  }
}

provider "mtncloud" {
  token = var.mtncloud_token
}

variable "mtncloud_token" {
  type      = string
  sensitive = true
}

resource "mtncloud_security_group" "web" {
  name        = "web-servers"
  description = "Allow SSH and HTTPS"
}

resource "mtncloud_security_group_rule" "ssh" {
  security_group_id = mtncloud_security_group.web.id
  name              = "allow-ssh"
  direction         = "ingress"
  policy            = "accept"
  protocol          = "tcp"
  port_range        = "22"
  source_type       = "cidr"
  source            = "203.0.113.10/32"
  destination_type  = "instance"
  ethertype         = "IPv4"
  enabled           = true
}
