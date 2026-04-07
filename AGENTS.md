# AGENTS.md

## Project Overview
- **Module**: `gke-cost-advisor` (Go 1.22.2)
- **Type**: HTTP server for GKE cost optimization recommendations
- **Entry point**: `cmd/server/main.go`

## Architecture
- `cmd/server/` - Application entry point
- `internal/application/usecases/` - Use cases (business logic orchestration)
- `internal/domain/services/` - Domain services (EfficiencyScorer, RightSizer)
- `internal/domain/entities/`, `valueobjects/` - Domain models
- `internal/infrastructure/clients/` - External clients (Prometheus, Kubernetes, GCP Billing)
- `internal/infrastructure/repositories/` - Data access layer
- `internal/presentation/handlers/` - HTTP handlers
- `pkg/config/`, `pkg/errors/` - Shared packages

## Implemented Use Cases
- `AnalyzeCostEfficiencyUseCase` - Namespace-level cost efficiency scoring
- `GenerateRightSizingRecommendationsUseCase` - CPU/memory recommendations
- `CalculateFinancialImpactUseCase` - Financial savings calculation
- `CompareMachineFamiliesUseCase` - Machine family comparison

## API Endpoints
- `GET /api/v1/clusters` - List clusters
- `GET /api/v1/clusters/{id}/efficiency` - Get efficiency scores
- `GET /api/v1/recommendations` - List recommendations
- `POST /api/v1/machine-families/compare` - Compare machine families
- `GET /health` - Health check

## Running the Server
```bash
go run cmd/server/main.go
```

## Environment Variables
- `SERVER_PORT` - HTTP server port (default: 8080)
- `PROMETHEUS_ENDPOINT` - Prometheus URL (default: http://localhost:9090)
- `PROMETHEUS_TOKEN` - Optional auth token
- `KUBECONFIG` - Path to kubeconfig file
- `GCP_PROJECT_ID` - GCP project ID
- `GCP_REGION` - GCP region (default: us-central1)

Dependencies managed via `go.mod` - run `go mod tidy` after adding imports.

## Development Notes
- No tests exist yet (`**/*_test.go` returns no matches)
- No CI/CD workflows or Makefile in root
- Kubernetes client falls back to stub data when no kubeconfig available
- Pricing data is static (E2, N2, C3 machine families)

## OpenCode Context
- `.opencode/memory/project.md` - Project memory block
- `.opencode/agents/` - Contains agent definitions (backend, frontend, etc.)
- Master plan: `GKE_Cost_Advisor_Platform_Master_Plan.md`
