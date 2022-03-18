Reference architecture for a deployment with servers running in AWS Fargate.

## Prerequisites
- You must have a route53 public hosted zone, passed in via the `zone_name`.  This is where the certificate verification records and the CNAME of the `service_domain_name` will be made.


## Deploying

To run, make a `my-deployment.tfvars` file and put your inputs in to it. 

```terraform
region              = "us-west-2"
availability_zones  = ["us-west-2c", "us-west-2b", "us-west-2a"]
zone_name           = "test.example.com"
service_domain_name = "edgeproxy.test.example.xyz"
```

Then apply 

```shell
terraform apply -var-file my-deployment.tfvars
```

# Terraform

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | 4.5.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 4.5.0 |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_base-network"></a> [base-network](#module\_base-network) | cn-terraform/networking/aws | n/a |
| <a name="module_ecs-fargate"></a> [ecs-fargate](#module\_ecs-fargate) | cn-terraform/ecs-fargate/aws | 2.0.34 |

## Resources

| Name | Type |
|------|------|
| [aws_acm_certificate.this](https://registry.terraform.io/providers/hashicorp/aws/4.5.0/docs/resources/acm_certificate) | resource |
| [aws_acm_certificate_validation.this](https://registry.terraform.io/providers/hashicorp/aws/4.5.0/docs/resources/acm_certificate_validation) | resource |
| [aws_cloudwatch_log_group.logs](https://registry.terraform.io/providers/hashicorp/aws/4.5.0/docs/resources/cloudwatch_log_group) | resource |
| [aws_route53_record.example](https://registry.terraform.io/providers/hashicorp/aws/4.5.0/docs/resources/route53_record) | resource |
| [aws_route53_record.service](https://registry.terraform.io/providers/hashicorp/aws/4.5.0/docs/resources/route53_record) | resource |
| [aws_route53_zone.default](https://registry.terraform.io/providers/hashicorp/aws/4.5.0/docs/data-sources/route53_zone) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_availability_zones"></a> [availability\_zones](#input\_availability\_zones) | list of AZes within the region to use.  Must be at least 2, i.e ["us-west-2a", "us-west-2-b"] | `list(string)` | n/a | yes |
| <a name="input_edgeproxy_image"></a> [edgeproxy\_image](#input\_edgeproxy\_image) | where to source the server image from | `string` | `"ghcr.io/segator/edgeproxy"` | no |
| <a name="input_edgeproxy_tag"></a> [edgeproxy\_tag](#input\_edgeproxy\_tag) | the server image tag to use | `string` | `"main"` | no |
| <a name="input_region"></a> [region](#input\_region) | AWS region to deploy this server | `any` | n/a | yes |
| <a name="input_service_domain_name"></a> [service\_domain\_name](#input\_service\_domain\_name) | FQDN of our service. i.e. "edgeproxy.test.example.com" | `any` | n/a | yes |
| <a name="input_zone_name"></a> [zone\_name](#input\_zone\_name) | domain name of the r53 zone records will go in, i.e. "test.example.com" | `any` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_aws_lb_lb_dns_name"></a> [aws\_lb\_lb\_dns\_name](#output\_aws\_lb\_lb\_dns\_name) | n/a |
<!-- END_TF_DOCS -->