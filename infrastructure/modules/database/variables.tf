# infrastructure/modules/database/variables.tf

variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_cidrs" { type = list(string) }
variable "private_subnet_ids" { type = list(string) }