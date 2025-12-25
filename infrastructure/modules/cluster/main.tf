# infrastructure/modules/cluster/main.tf

# 1. Get the latest Ubuntu Image
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

# 2. Create Security Group for the Cluster
resource "aws_security_group" "k3s_sg" {
  name        = "${var.environment}-k3s-node-sg"
  description = "Allow inbound traffic for K3s"
  vpc_id      = var.vpc_id

  # --- NEW RULE: HTTP Web Traffic (The Fix) ---
  ingress {
    description = "Allow HTTP Web Traffic"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # SSH Access (For you to connect)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] 
  }

  # Kubernetes API Access (For kubectl)
  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # App Traffic (NodePorts 30000-32767)
  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# 3. The EC2 Instance (Your Server)
resource "aws_instance" "k3s_node" {
  ami           = data.aws_ami.ubuntu.id
  
  # --- UPDATED: Matches your current upgrade ---
  instance_type = "t3.small" 
  
  subnet_id     = var.public_subnet_id
  key_name      = var.key_name
  
  vpc_security_group_ids = [aws_security_group.k3s_sg.id]

  # This script runs ONCE when the server starts
  user_data = <<-EOF
              #!/bin/bash
              # Install K3s (Lightweight Kubernetes)
              curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644
              
              # Wait for K3s to start
              sleep 15
              
              # Install Helm (Package Manager)
              curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
              EOF

  tags = {
    Name = "${var.environment}-k3s-cluster"
  }
}