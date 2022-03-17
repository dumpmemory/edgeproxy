variable "region" {}

variable "availability_zones" {
  type = list(string)
}

variable "edgeproxy_image" {
  default = "ghcr.io/djcrabhat/edgeproxy"
}

variable "edgeproxy_tag" {
  default = "feature-aws-reference-arch"
}

variable "zone_name" {
  description = "name of the r53 zone records will go in"
}

variable "service_domain_name" {}