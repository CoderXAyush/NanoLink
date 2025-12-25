# infrastructure/environments/dev/terraform.tfvars

environment = "dev"
cluster_name = "url-shortener"

# The main IP range for your entire network
vpc_cidr = "10.0.0.0/16" 

# Availability Zones (Must match your region, e.g., us-east-1a)
availability_zones = ["us-east-1a", "us-east-1b"]

# Public Subnets (Load Balancers go here)
public_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24"]

# Private Subnets (Apps & DBs go here - safer!)
private_subnet_cidrs = ["10.0.3.0/24", "10.0.4.0/24"]

# The SSH Key you created in AWS Console (MUST MATCH EXACTLY)
key_name = "devops-project-key"