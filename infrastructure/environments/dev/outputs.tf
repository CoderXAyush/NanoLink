# infrastructure/environments/dev/outputs.tf

output "k3s_public_ip" {
  description = "Public IP of the Kubernetes Cluster"
  value       = module.cluster.k3s_public_ip
}

output "redis_endpoint" {
  description = "The address of the Redis Cache"
  value       = module.database.redis_endpoint
}