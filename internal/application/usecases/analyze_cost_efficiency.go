package usecases

import (
	"context"
	"time"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/services"
	"gke-cost-advisor/internal/domain/valueobjects"
	"gke-cost-advisor/pkg/errors"
)

// AnalyzeCostEfficiencyRequest represents a request to analyze cost efficiency
type AnalyzeCostEfficiencyRequest struct {
	ClusterID    string
	Namespaces   []string
	WindowConfig valueobjects.WindowConfig
}

// AnalyzeCostEfficiencyResponse represents the response for cost efficiency analysis
type AnalyzeCostEfficiencyResponse struct {
	ClusterID       string
	ClusterName     string
	Region          string
	CompositeScore  float64
	ColorCode       valueobjects.ScoreColor
	NamespaceScores []NamespaceScoreResult
	AnalyzedAt      time.Time
	DataFreshness   time.Duration
}

// NamespaceScoreResult represents namespace-level score
type NamespaceScoreResult struct {
	Namespace         string
	Score             float64
	ColorCode         valueobjects.ScoreColor
	CPUUtilization    float64
	MemoryUtilization float64
}

// ClusterRepository provides cluster data access methods
type ClusterRepository interface {
	GetByID(ctx context.Context, id string) (*entities.Cluster, error)
	List(ctx context.Context) ([]*entities.Cluster, error)
	ListNamespaces(ctx context.Context, clusterID string) ([]string, error)
}

// MetricsRepository provides metrics data access methods
type MetricsRepository interface {
	GetNamespaceMetrics(ctx context.Context, clusterID, namespace string, timeRange valueobjects.TimeRange) (services.NamespaceMetrics, error)
}

// AnalyzeCostEfficiencyUseCase implements the cost efficiency analysis use case
type AnalyzeCostEfficiencyUseCase struct {
	clusterRepo ClusterRepository
	metricsRepo MetricsRepository
	scorer      services.EfficiencyScorer
}

// NewAnalyzeCostEfficiencyUseCase creates a new use case
func NewAnalyzeCostEfficiencyUseCase(
	clusterRepo ClusterRepository,
	metricsRepo MetricsRepository,
	scorer services.EfficiencyScorer,
) *AnalyzeCostEfficiencyUseCase {
	return &AnalyzeCostEfficiencyUseCase{
		clusterRepo: clusterRepo,
		metricsRepo: metricsRepo,
		scorer:      scorer,
	}
}

// Execute performs the cost efficiency analysis
func (uc *AnalyzeCostEfficiencyUseCase) Execute(ctx context.Context, req AnalyzeCostEfficiencyRequest) (*AnalyzeCostEfficiencyResponse, error) {
	if req.ClusterID == "" {
		return nil, errors.NewApplicationError("cluster ID is required", errors.ErrInvalidInput)
	}

	cluster, err := uc.clusterRepo.GetByID(ctx, req.ClusterID)
	if err != nil {
		return nil, errors.NewApplicationError("cluster not found", errors.ErrClusterNotFound)
	}

	namespaces := req.Namespaces
	if len(namespaces) == 0 {
		nsList, err := uc.clusterRepo.ListNamespaces(ctx, req.ClusterID)
		if err != nil {
			return nil, err
		}
		namespaces = nsList
	}

	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-req.WindowConfig.Size), now)

	namespaceScores := make([]NamespaceScoreResult, 0, len(namespaces))
	var totalCPU, totalMemory, totalStorage float64

	for _, ns := range namespaces {
		metrics, err := uc.metricsRepo.GetNamespaceMetrics(ctx, req.ClusterID, ns, timeRange)
		if err != nil {
			continue
		}

		score := uc.scorer.CalculateNamespaceScore(metrics)
		namespaceScores = append(namespaceScores, NamespaceScoreResult{
			Namespace:         ns,
			Score:             score.CompositeScore,
			ColorCode:         score.ColorCode,
			CPUUtilization:    score.CPUUtilization,
			MemoryUtilization: score.MemoryUtilization,
		})

		totalCPU += score.CPUUtilization
		totalMemory += score.MemoryUtilization
		totalStorage += score.StorageEfficiency
	}

	count := float64(len(namespaceScores))
	if count == 0 {
		return &AnalyzeCostEfficiencyResponse{
			ClusterID:      cluster.ID,
			ClusterName:    cluster.Name,
			Region:         cluster.Region,
			CompositeScore: 0,
			ColorCode:      valueobjects.ScoreColorRed,
			AnalyzedAt:     now,
		}, nil
	}

	avgCPU := totalCPU / count
	avgMemory := totalMemory / count
	avgStorage := totalStorage / count

	compositeScore := (avgCPU * 0.30) + (avgMemory * 0.40) + (avgStorage * 0.30)

	colorCode := valueobjects.ScoreColorRed
	if compositeScore >= 80 {
		colorCode = valueobjects.ScoreColorGreen
	} else if compositeScore >= 50 {
		colorCode = valueobjects.ScoreColorYellow
	}

	return &AnalyzeCostEfficiencyResponse{
		ClusterID:       cluster.ID,
		ClusterName:     cluster.Name,
		Region:          cluster.Region,
		CompositeScore:  compositeScore,
		ColorCode:       colorCode,
		NamespaceScores: namespaceScores,
		AnalyzedAt:      now,
		DataFreshness:   5 * time.Minute,
	}, nil
}
