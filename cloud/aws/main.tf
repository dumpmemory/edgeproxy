terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

resource "aws_cloudwatch_log_group" "logs" {
  name_prefix = "edgeproxy"
}


module "base-network" {
  source                                      = "cn-terraform/networking/aws"
  name_prefix                                 = "edgeproxy-example-service"
  vpc_cidr_block                              = "192.168.0.0/16"
  availability_zones                          = var.availability_zones
  public_subnets_cidrs_per_availability_zone  = ["192.168.0.0/19", "192.168.32.0/19", "192.168.64.0/19", "192.168.96.0/19"]
  private_subnets_cidrs_per_availability_zone = ["192.168.128.0/19", "192.168.160.0/19", "192.168.192.0/19", "192.168.224.0/19"]
}

module "ecs-fargate" {
  source  = "cn-terraform/ecs-fargate/aws"
  version = "2.0.34"

  name_prefix     = "edgeproxy-exmpl-"
  vpc_id          = module.base-network.vpc_id
  container_image = "${var.edgeproxy_image}:${var.edgeproxy_tag}"
  container_name  = "edgeproxy"
  command         = ["server", "--verbose"]

  public_subnets_ids  = module.base-network.public_subnets_ids
  private_subnets_ids = module.base-network.private_subnets_ids

  default_certificate_arn = aws_acm_certificate.this.arn
  port_mappings = [
    {
      "containerPort" : 9180,
      hostPort : 9180,
      "protocol" : "tcp"
    }
  ]

  lb_target_group_health_check_path = "/readyz"
  lb_https_ports = {
    "default_http" : {
      "listener_port" : 443,
      "target_group_port" : 9180,
      "target_group_protocol" : "HTTP"
    }
  }
  lb_http_ports = {}

  log_configuration = {
    "logDriver" : "awslogs",
    "options" : {
      "awslogs-group" : "${aws_cloudwatch_log_group.logs.name}",
      "awslogs-region" : var.region,
      "awslogs-stream-prefix" : "streaming"
    }
  }
}



resource "aws_route53_record" "service" {

  allow_overwrite = true
  name            = var.service_domain_name
  records         = [module.ecs-fargate.aws_lb_lb_dns_name]
  ttl             = 60
  type            = "CNAME"
  zone_id         = data.aws_route53_zone.default.zone_id
}
output "aws_lb_lb_dns_name" {
  value = module.ecs-fargate.aws_lb_lb_dns_name
}