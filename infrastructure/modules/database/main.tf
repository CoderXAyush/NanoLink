# infrastructure/modules/database/main.tf

# 1. Security Group for Redis
# This acts as a firewall. It says "Only allow traffic from the App, not the internet."
resource "aws_security_group" "redis_sg" {
  name        = "${var.environment}-redis-sg"
  description = "Allow inbound traffic from private subnets"
  vpc_id      = var.vpc_id

  # Inbound Rule: Allow TCP port 6379 (Redis)
  ingress {
    description = "Redis from VPC"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = var.private_subnet_cidrs # Only apps in private subnets can talk to Redis
  }

  # Outbound Rule: Allow everything
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.environment}-redis-sg"
  }
}

# 2. Redis Subnet Group
# Tells AWS "Put this Redis cluster inside these specific private subnets"
resource "aws_elasticache_subnet_group" "redis_subnet_group" {
  name       = "${var.environment}-redis-subnet-group"
  subnet_ids = var.private_subnet_ids
}

# 3. The Redis Cluster (ElastiCache)
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "${var.environment}-redis-cluster"
  engine               = "redis"
  node_type            = "cache.t3.micro" # Smallest size for testing
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  engine_version       = "7.0"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.redis_subnet_group.name
  security_group_ids   = [aws_security_group.redis_sg.id]
}