# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    aws      = "3.3.0"
    ct       = "0.6.1"
    null     = "2.1.2"
    template = "2.1.2"
    random   = "2.3.0"
  }
}
