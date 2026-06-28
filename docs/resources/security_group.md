# mtncloud_security_group

Manages an MTN Cloud security group.

```hcl
resource "mtncloud_security_group" "web" {
  name        = "web-servers"
  description = "Allow SSH and HTTPS"
}
```

Import:

```bash
terraform import mtncloud_security_group.web 42
```
