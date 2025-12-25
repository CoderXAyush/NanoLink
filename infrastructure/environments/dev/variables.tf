# We re-declare these here so we can pass them from .tfvars to the module

variable "environment" {}
variable "vpc_cidr" {}
variable "availability_zones" {}
variable "public_subnet_cidrs" {}
variable "private_subnet_cidrs" {}
variable "cluster_name" {}
variable "key_name" {}