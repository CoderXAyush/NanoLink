provider "aws" {
  region = "us-east-1"
}

resource "aws_security_group" "nanolink_sg" {
  name        = "nanolink-security-group"
  description = "Allow HTTP and API traffic"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "app_server" {
  ami           = "ami-0c7217cdde317cfec" # Ubuntu 22.04 Free Tier
  instance_type = "t2.micro"
  security_groups = [aws_security_group.nanolink_sg.name]

  tags = {
    Name = "NanoLink-Server"
  }
}
