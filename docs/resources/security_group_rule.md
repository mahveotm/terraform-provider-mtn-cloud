# mtncloud_security_group_rule

Manages a rule within an MTN Cloud security group.

```hcl
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
```

Import:

```bash
terraform import mtncloud_security_group_rule.ssh 42:99
```
