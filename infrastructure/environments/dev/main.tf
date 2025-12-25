terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# 1. Build the Network (VPC, Subnets, Internet Gateway)
module "networking" {
  source = "../../modules/networking"

  environment          = var.environment
  vpc_cidr             = var.vpc_cidr
  availability_zones   = var.availability_zones
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  cluster_name         = var.cluster_name
}

# 2. Build the Database (Redis Cache)
module "database" {
  source = "../../modules/database"

  environment          = var.environment
  vpc_id               = module.networking.vpc_id
  private_subnet_cidrs = var.private_subnet_cidrs
  private_subnet_ids   = module.networking.private_subnet_ids
}

# 3. Build the Cluster (The Server)
module "cluster" {
  source = "../../modules/cluster"

  environment      = var.environment
  vpc_id           = module.networking.vpc_id
  
  # Put the server in the FIRST Public Subnet so it can reach the internet
  public_subnet_id = module.networking.public_subnet_ids[0]
  
  # The SSH Key you created in AWS Console
  key_name         = var.key_name
}