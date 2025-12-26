# 🚀 NanoLink: High-Performance Distributed URL Shortener

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![Terraform](https://img.shields.io/badge/Terraform-1.5+-7B42BC?style=flat&logo=terraform)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Kustomize-326CE5?style=flat&logo=kubernetes)
![AWS](https://img.shields.io/badge/AWS-Infrastructure-232F3E?style=flat&logo=amazon-aws)
![Redis](https://img.shields.io/badge/Redis-Caching-DC382D?style=flat&logo=redis)
![Kafka](https://img.shields.io/badge/Kafka-Streaming-231F20?style=flat&logo=apachekafka)

**NanoLink** is a production-grade, distributed URL shortening service designed to handle **Twitter-scale traffic**. 

Unlike basic CRUD apps, this system solves distributed concurrency challenges like the **Thundering Herd problem**, **Unique ID generation across replicas**, and **Non-blocking analytics**. It demonstrates a complete DevOps lifecycle using a Monorepo pattern.

---

## 🛠️ Tech Stack Used

| Technology | Role | Reason |
| :--- | :--- | :--- |
| **Go (Golang)** | Backend | **Concurrency:** Go's lightweight Goroutines allow us to handle 10k+ concurrent requests per second with minimal RAM, perfect for the high-throughput redirection service. |
| **Terraform** | IaC | **Reproducibility:** We avoid "ClickOps" by defining all AWS resources (VPC, EKS, RDS) as code. This ensures our Dev and Prod environments are identical. |
| **Kubernetes** | Orchestration | **Autoscaling:** K8s monitors CPU usage. If traffic spikes (e.g., a viral link), the Horizontal Pod Autoscaler (HPA) automatically spins up new pods to handle the load. |
| **AWS** | Cloud Provider | **Managed Services:** We leverage managed EKS (for K8s) and ElastiCache (for Redis) to reduce operational overhead and focus on application logic. |
| **Redis** | Caching & State | **Performance:** Used for two critical functions: <br>1. **Caching** hot URLs to ensure <10ms response times.<br>2. **Atomic Counters** to reserve ID blocks for distributed unique ID generation. |
| **Kafka** | Message Queue | **Decoupling:** Prevents database lockups. By offloading click tracking to Kafka, the user gets redirected instantly while we process analytics safely in the background. |

## 🏗️ Architecture

*(Note: System architecture featuring separation of Read/Write paths)*

The system is split into decoupled microservices to allow independent scaling:
* **Shortener Service (Write):** Handles ID generation and URL mapping.
* **Redirect Service (Read):** High-performance redirection optimized for low latency (<10ms).
* **Analytics Worker:** Background consumer processing click data without blocking user responses.

---

## 📂 Repository Structure

This repository follows a strict Monorepo pattern, separating application logic from infrastructure.

```text
├── application/         # 🧠 Source Code
│   ├── cmd/             # Entrypoints (api-server, analytics-worker)
│   ├── internal/        # Private library code (core logic, handlers)
│   └── Dockerfile       # Multi-stage build for containers
├── infrastructure/      # 🏗️ Infrastructure as Code (Terraform)
│   ├── modules/         # Reusable modules (Networking, EKS, RDS)
│   └── environments/    # Environment-specific state (Dev, Prod)
└── k8s/                 # ☸️ Kubernetes Manifests
    ├── base/            # Common deployment logic
    └── overlays/        # Kustomize patches (Dev vs Prod)

    

## 🔄 Application Workflow

### 1. URL Shortening (Write Path)

1. **Request:** User sends a long URL to the API.
2. **ID Allocation:** API Server reserves a unique ID block from **Redis**.
3. **Encoding:** Server converts the ID to a Base62 string (e.g., `Ab3d9X`).
4. **Persistence:** Mapping is saved to **Cassandra/Mongo** for durability.
5. **Response:** Short URL is returned to the user.

### 2. Redirection (Read Path - High Speed)

1. **Access:** User clicks the short link.
2. **Cache Lookup:** API Server checks **Redis**.
* *Hit:* Returns 301 Redirect immediately.
* *Miss:* Fetches from DB, updates Redis, then redirects.


3. **Async Logging:** Server pushes "Click Event" to **Kafka** (Non-blocking).

### 3. Analytics Processing (Background Path)

1. **Ingestion:** **Analytics Worker** consumes Kafka topic.
2. **Aggregation:** Buffers events and processes in batches.
3. **Storage:** Stats written to **Time-Series Database**.

---

## 🚀 Getting Started

### Prerequisites

* **Go**: 1.22+
* **Docker**: For containerization
* **Terraform**: 1.5+
* **AWS CLI**: Configured
* **Kubectl**: Connected to cluster

### 1. Backend Development (Local)

Navigate to the application directory to run services:

```bash
cd application

# Run API Server
go run cmd/api-server/main.go

# Run Analytics Worker
go run cmd/analytics-worker/main.go

```

### 2. Build & Run (Docker)

Build the unified container image:

```bash
# Build from root context
docker build -f application/Dockerfile -t nanolink:local .

# Run container
docker run -p 8080:8080 nanolink:local

```

---

## ☁️ Infrastructure (Terraform)

Infrastructure is managed with strict environment isolation.

### Provisioning Dev Environment

```bash
cd infrastructure/environments/dev

# 1. Initialize
terraform init

# 2. Plan
terraform plan -out=tfplan

# 3. Apply
terraform apply tfplan

```

> **Note:** Never commit `.tfvars` or `.tfstate` files containing credentials.

## ☸️ Kubernetes Deployment

I use **Kustomize** for environment-specific configuration management.

### Deploy to Cluster

```bash
# Preview Manifests
kubectl kustomize k8s/overlays/dev

# Apply to Cluster
kubectl apply -k k8s/overlays/dev

```

### Rolling Updates

Update the image without editing YAML:

```bash
cd k8s/overlays/dev
kustomize edit set image url-shortener-image=registry/nanolink:v2.0.1
kubectl apply -k .

```

---

## 🧪 Testing Strategy

### Unit Tests

```bash
cd application
go test ./... -v

```

### Load Testing (k6)

I use **k6** to simulate high concurrency.

```bash
# Install k6 and run script
k6 run scripts/load_test.js

```

*Target Metrics:* > 2,000 RPS, < 50ms Latency, < 1% Error Rate.
