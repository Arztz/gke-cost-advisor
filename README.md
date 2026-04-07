# GKE Cost Advisor

A Go-based HTTP server that provides cost optimization recommendations for Google Kubernetes Engine (GKE) clusters. The platform analyzes cluster efficiency, generates right-sizing recommendations, and helps compare machine families to optimize cloud spending.

## Overview

GKE Cost Advisor integrates with:
- **Prometheus** - For container metrics (CPU, memory, throttling)
- **Kubernetes** - For cluster, node pool, and container information
- **GCP Billing** - For machine type pricing data

The platform provides:
- **Cost Efficiency Scoring** (0-100 scale with color coding)
- **Right-Sizing Recommendations** based on P95 utilization metrics
- **Machine Family Comparison** to find cost-effective alternatives
- **Financial Impact Analysis** for estimated savings

## Quick Start

### Prerequisites

- Go 1.22.2 or later
- Access to a GKE cluster with Prometheus metrics
- Kubernetes configuration (optional, uses stub data if unavailable)

### Run the Server

```bash
# Clone and navigate to the project
cd gke-cost-advisor

# Run the server
go run cmd/server/main.go
```

The server starts on port 8080 by default.

### Run with Docker

```bash
# Build the Docker image
docker build -t gke-cost-advisor:latest .

# Run the container
docker run -p 8080:8080 \
  -e PROMETHEUS_ENDPOINT=http://prometheus:9090 \
  -e GCP_PROJECT_ID=my-project \
  -e GCP_REGION=us-central1 \
  gke-cost-advisor:latest
```

## Configuration

Configure the server using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `PROMETHEUS_ENDPOINT` | Prometheus URL | `http://localhost:9090` |
| `PROMETHEUS_TOKEN` | Optional auth token | (empty) |
| `PROMETHEUS_TIMEOUT` | Query timeout | `30s` |
| `KUBECONFIG` | Path to kubeconfig file | (in-cluster or ~/.kube/config) |
| `KUBERNETES_NAMESPACE` | Target namespace | (all namespaces) |
| `GCP_PROJECT_ID` | GCP project ID | (empty) |
| `GCP_BILLING_ACCOUNT_ID` | Billing account ID | (empty) |
| `GCP_REGION` | GCP region | `us-central1` |
| `ANALYSIS_DEFAULT_WINDOW` | Analysis time window | `24h` |
| `ANALYSIS_CPU_HEADROOM` | CPU headroom percentage | `0.20` |
| `ANALYSIS_MEMORY_HEADROOM` | Memory headroom percentage | `0.30` |

## API Reference

### Health Check

Check if the server is running.

```bash
GET /health
```

**Response:**
```json
{"status":"ok"}
```

---

### List Clusters

Get all available GKE clusters.

```bash
GET /api/v1/clusters
```

**Response:**
```json
{
  "clusters": [
    {
      "id": "cluster-1",
      "name": "production-cluster",
      "region": "us-central1",
      "location": "us-central1-a",
      "project_id": "my-project",
      "master_version": "1.28.3-gke.1700",
      "node_pools": [...],
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

### Get Cluster Efficiency

Analyze cost efficiency for a specific cluster.

```bash
GET /api/v1/clusters/{id}/efficiency
```

**Parameters:**
- `id` (path) - Cluster ID

**Response:**
```json
{
  "cluster_id": "cluster-1",
  "namespace_scores": [
    {
      "namespace": "production",
      "score": {
        "composite_score": 72.5,
        "cpu_utilization": 65.0,
        "memory_utilization": 70.0,
        "storage_efficiency": 82.0,
        "time_range": {
          "start_time": "2024-01-20T00:00:00Z",
          "end_time": "2024-01-21T00:00:00Z",
          "duration": "24h0m0s"
        },
        "confidence": "high",
        "color_code": "YELLOW"
      }
    }
  ],
  "overall_score": {
    "composite_score": 72.5,
    "cpu_utilization": 65.0,
    "memory_utilization": 70.0,
    "storage_efficiency": 82.0,
    "confidence": "high",
    "color_code": "YELLOW"
  }
}
```

**Score Interpretation:**
- `GREEN` (80-100): Efficient resource usage
- `YELLOW` (50-79): Room for improvement
- `RED` (0-49): Significant optimization opportunity

The composite score uses a weighted formula: **30% CPU + 40% Memory + 30% Storage**

---

### Get Recommendations

Retrieve right-sizing recommendations for a cluster.

```bash
GET /api/v1/recommendations?cluster_id={cluster_id}
```

**Parameters:**
- `cluster_id` (query) - Required cluster ID

**Response:**
```json
{
  "recommendations": [
    {
      "id": "rec-container-123",
      "cluster_id": "cluster-1",
      "recommendation_type": "right_sizing",
      "target_type": "container",
      "target_id": "container-123",
      "target_name": "api-server",
      "current_value": "current",
      "recommended_value": "recommended",
      "confidence": "medium",
      "justification": "Resource optimization opportunity detected",
      "status": "pending",
      "created_at": "2024-01-21T12:00:00Z",
      "updated_at": "2024-01-21T12:00:00Z"
    }
  ],
  "total": 1
}
```

**Recommendation Types:**
- `right_sizing`: Adjust CPU/memory requests
- `spot_migration`: Migrate to Spot instances
- `node_pool`: Optimize node pool configuration
- `limit_adjust`: Adjust resource limits

---

### Compare Machine Families

Compare machine families based on resource requirements.

```bash
POST /api/v1/machine-families/compare
```

**Request Body:**
```json
{
  "requirements": {
    "vcpu": 4,
    "memory_gb": 16,
    "storage_gb": 100
  },
  "region": "us-central1",
  "include_spot": true,
  "latency_sensitive": false
}
```

**Response:**
```json
{
  "comparisons": [
    {
      "machine_family": "e2",
      "machine_type": "e2-standard-4",
      "on_demand_hourly": 0.192,
      "spot_hourly": 0.058,
      "spot_savings_percent": 69.8,
      "performance_score": 85,
      "price_performance_ratio": 2.5,
      "recommended": true
    },
    {
      "machine_family": "n2",
      "machine_type": "n2-standard-4",
      "on_demand_hourly": 0.224,
      "spot_hourly": 0.067,
      "spot_savings_percent": 70.1,
      "performance_score": 90,
      "price_performance_ratio": 2.3,
      "recommended": false
    }
  ]
}
```

## Architecture

The project follows **Clean Architecture** with Domain-Driven Design (DDD) principles:

```
cmd/server/
  └── main.go                 # Entry point

internal/
  ├── application/
  │   └── usecases/          # Business logic orchestration
  ├── domain/
  │   ├── entities/          # Domain models (Cluster, Container, Workload)
  │   ├── services/          # Domain services (EfficiencyScorer, RightSizer)
  │   └── valueobjects/      # Value objects (CostEfficiencyScore, Savings)
  ├── infrastructure/
  │   ├── clients/           # External integrations (Prometheus, K8s, GCP)
  │   └── repositories/      # Data access layer
  └── presentation/
      └── handlers/          # HTTP handlers

pkg/
  ├── config/               # Configuration management
  └── errors/               # Error handling
```

### Key Components

- **EfficiencyScorer**: Calculates composite scores (0-100) using weighted metrics
- **RightSizer**: Generates resource recommendations based on P95 utilization
- **PrometheusClient**: Queries container metrics via PromQL
- **KubernetesClient**: Retrieves cluster, pod, and container information
- **BillingClient**: Provides machine type pricing data

## Examples

### Example 1: Check Cluster Efficiency

```bash
# Start the server
go run cmd/server/main.go &

# Get efficiency scores for a cluster
curl http://localhost:8080/api/v1/clusters/cluster-1/efficiency
```

Expected output shows composite score and per-namespace breakdown with color coding.

### Example 2: Get Optimization Recommendations

```bash
# Get right-sizing recommendations
curl "http://localhost:8080/api/v1/recommendations?cluster_id=cluster-1"
```

Returns containers with CPU or memory waste > 10% based on P95 metrics.

### Example 3: Compare Machine Families

```bash
# Compare machines for a workload requiring 4 vCPU, 16GB RAM
curl -X POST http://localhost:8080/api/v1/machine-families/compare \
  -H "Content-Type: application/json" \
  -d '{
    "requirements": {"vcpu": 4, "memory_gb": 16, "storage_gb": 100},
    "region": "us-central1",
    "include_spot": true
  }'
```

Shows on-demand and spot pricing with savings percentages.

### Example 4: Run with Docker Compose

```yaml
version: '3.8'
services:
  gke-cost-advisor:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SERVER_PORT=8080
      - PROMETHEUS_ENDPOINT=http://prometheus:9090
      - GCP_PROJECT_ID=my-project
      - GCP_REGION=us-central1
    depends_on:
      - prometheus

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

## Development

See [AGENTS.md](./AGENTS.md) for development details, including:

- Implemented use cases
- Unit testing
- API endpoints
- Running tests

### Run Tests

```bash
go test ./...
```

### Build Binary

```bash
go build -o server ./cmd/server
```

## License

MIT License
