terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "4.5.0"
    }
  }
}

provider "aws" {
  region = var.region
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

  name_prefix         = "edgeproxy-example-service"
  vpc_id              = module.base-network.vpc_id
  container_image     = "ubuntu"
  container_name      = "test"
  public_subnets_ids  = module.base-network.public_subnets_ids
  private_subnets_ids = module.base-network.private_subnets_ids
}