variable "region" {
  description = "AWS region to deploy this server"
}

variable "availability_zones" {
  type = list(string)
  description = "list of AZes within the region to use.  Must be at least 2, i.e [\"us-west-2a\", \"us-west-2-b\"]"
}

variable "edgeproxy_image" {
  default = "ghcr.io/segator/edgeproxy"
  description = "where to source the server image from"
}

variable "edgeproxy_tag" {
  default = "main"
  description = "the server image tag to use"
}

variable "zone_name" {
  description = "domain name of the r53 zone records will go in, i.e. \"test.example.com\""
}

variable "service_domain_name" {
  description = "FQDN of our service. i.e. \"edgeproxy.test.example.com\""
}