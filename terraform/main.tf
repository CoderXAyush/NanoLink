provider "aws" {
  region = "ap-south-1"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = "url-shortener-vpc"
  }
}

resource "aws_ecr_repository" "repo" {
  name = "url-shortener"
}
