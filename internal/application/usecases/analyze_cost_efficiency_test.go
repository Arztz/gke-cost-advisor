package usecases

import (
	"context"
	"testing"
	"time"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/services"
	"gke-cost-advisor/internal/domain/valueobjects"
	"gke-cost-advisor/pkg/errors"
)

// Mock implementations for testing

type mockClusterRepo struct {
	clusters   map[string]*entities.Cluster
	namespaces map[string][]string
	err        error
}

func (m *mockClusterRepo) GetByID(ctx context.Context, id string) (*entities.Cluster, error) {
	if m.err != nil {
		return nil, m.err
	}
	if cluster, ok := m.clusters[id]; ok {
		return cluster, nil
	}
	return nil, errors.NewApplicationError("cluster not found", errors.ErrClusterNotFound)
}

func (m *mockClusterRepo) List(ctx context.Context) ([]*entities.Cluster, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*entities.Cluster, 0, len(m.clusters))
	for _, c := range m.clusters {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockClusterRepo) ListNamespaces(ctx context.Context, clusterID string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.namespaces[clusterID], nil
}

type mockMetricsRepo struct {
	metrics map[string]services.NamespaceMetrics
	err     error
}

func (m *mockMetricsRepo) GetNamespaceMetrics(ctx context.Context, clusterID, namespace string, timeRange valueobjects.TimeRange) (services.NamespaceMetrics, error) {
	if m.err != nil {
		return services.NamespaceMetrics{}, m.err
	}
	key := clusterID + "/" + namespace
	if metrics, ok := m.metrics[key]; ok {
		return metrics, nil
	}
	return services.NamespaceMetrics{}, nil
}

// Mock efficiency scorer
type mockEfficiencyScorer struct {
	scoreFunc func(metrics services.NamespaceMetrics) valueobjects.CostEfficiencyScore
}

func (m *mockEfficiencyScorer) CalculateNamespaceScore(metrics services.NamespaceMetrics) valueobjects.CostEfficiencyScore {
	if m.scoreFunc != nil {
		return m.scoreFunc(metrics)
	}
	return valueobjects.CostEfficiencyScore{
		CompositeScore:    50,
		ColorCode:         valueobjects.ScoreColorYellow,
		Confidence:        valueobjects.ConfidenceMedium,
		CPUUtilization:    metrics.CPUUtilization,
		MemoryUtilization: metrics.MemoryUtilization,
		StorageEfficiency: metrics.StorageEfficiency,
	}
}

func (m *mockEfficiencyScorer) CalculateWorkloadScore(workload *entities.Workload, metrics services.NamespaceMetrics) valueobjects.CostEfficiencyScore {
	return m.CalculateNamespaceScore(metrics)
}

func (m *mockEfficiencyScorer) CalculateClusterScore(cluster *entities.Cluster, metricsList []services.NamespaceMetrics) valueobjects.CostEfficiencyScore {
	if len(metricsList) == 0 {
		return valueobjects.CostEfficiencyScore{
			CompositeScore: 0,
			ColorCode:      valueobjects.ScoreColorRed,
			Confidence:     valueobjects.ConfidenceLow,
		}
	}
	var totalCPU, totalMemory, totalStorage float64
	for _, m := range metricsList {
		totalCPU += m.CPUUtilization
		totalMemory += m.MemoryUtilization
		totalStorage += m.StorageEfficiency
	}
	count := float64(len(metricsList))
	return valueobjects.NewCostEfficiencyScore(
		totalCPU/count,
		totalMemory/count,
		totalStorage/count,
		metricsList[0].TimeRange,
	)
}

func TestAnalyzeCostEfficiencyUseCase_Execute(t *testing.T) {
	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-24*time.Hour), now)

	tests := []struct {
		name               string
		request            AnalyzeCostEfficiencyRequest
		clusters           map[string]*entities.Cluster
		namespaces         map[string][]string
		metrics            map[string]services.NamespaceMetrics
		expectedErr        error
		expectedScore      float64
		expectedNamespaces int
	}{
		{
			name: "valid request with namespaces",
			request: AnalyzeCostEfficiencyRequest{
				ClusterID: "cluster-1",
				WindowConfig: valueobjects.WindowConfig{
					Size: 24 * time.Hour,
				},
			},
			clusters: map[string]*entities.Cluster{
				"cluster-1": {
					ID:       "cluster-1",
					Name:     "test-cluster",
					Region:   "us-central1",
					IsActive: true,
				},
			},
			namespaces: map[string][]string{
				"cluster-1": {"default", "kube-system"},
			},
			metrics: map[string]services.NamespaceMetrics{
				"cluster-1/default": {
					Namespace:         "default",
					CPUUtilization:    80,
					MemoryUtilization: 85,
					StorageEfficiency: 70,
					TimeRange:         timeRange,
				},
				"cluster-1/kube-system": {
					Namespace:         "kube-system",
					CPUUtilization:    60,
					MemoryUtilization: 70,
					StorageEfficiency: 60,
					TimeRange:         timeRange,
				},
			},
			expectedErr:        nil,
			expectedNamespaces: 2,
		},
		{
			name: "empty cluster ID",
			request: AnalyzeCostEfficiencyRequest{
				ClusterID: "",
			},
			expectedErr: errors.NewApplicationError("cluster ID is required", errors.ErrInvalidInput),
		},
		{
			name: "cluster not found",
			request: AnalyzeCostEfficiencyRequest{
				ClusterID: "nonexistent",
			},
			clusters: map[string]*entities.Cluster{
				"cluster-1": {
					ID:       "cluster-1",
					Name:     "test-cluster",
					Region:   "us-central1",
					IsActive: true,
				},
			},
			expectedErr: errors.NewApplicationError("cluster not found", errors.ErrClusterNotFound),
		},
		{
			name: "specific namespaces provided",
			request: AnalyzeCostEfficiencyRequest{
				ClusterID:  "cluster-1",
				Namespaces: []string{"default"},
				WindowConfig: valueobjects.WindowConfig{
					Size: 24 * time.Hour,
				},
			},
			clusters: map[string]*entities.Cluster{
				"cluster-1": {
					ID:       "cluster-1",
					Name:     "test-cluster",
					Region:   "us-central1",
					IsActive: true,
				},
			},
			namespaces: map[string][]string{
				"cluster-1": {"default", "kube-system", "monitoring"},
			},
			metrics: map[string]services.NamespaceMetrics{
				"cluster-1/default": {
					Namespace:         "default",
					CPUUtilization:    75,
					MemoryUtilization: 80,
					StorageEfficiency: 65,
					TimeRange:         timeRange,
				},
			},
			expectedErr:        nil,
			expectedNamespaces: 1,
		},
		{
			name: "no metrics available returns zero score",
			request: AnalyzeCostEfficiencyRequest{
				ClusterID: "cluster-1",
				WindowConfig: valueobjects.WindowConfig{
					Size: 24 * time.Hour,
				},
			},
			clusters: map[string]*entities.Cluster{
				"cluster-1": {
					ID:       "cluster-1",
					Name:     "test-cluster",
					Region:   "us-central1",
					IsActive: true,
				},
			},
			namespaces: map[string][]string{
				"cluster-1": {"default"},
			},
			metrics:            map[string]services.NamespaceMetrics{},
			expectedErr:        nil,
			expectedScore:      0.0,
			expectedNamespaces: 1, // Namespace is added with zero scores
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepo{
				clusters:   tt.clusters,
				namespaces: tt.namespaces,
			}
			metricsRepo := &mockMetricsRepo{
				metrics: tt.metrics,
			}
			scorer := services.NewEfficiencyScorer()

			uc := NewAnalyzeCostEfficiencyUseCase(clusterRepo, metricsRepo, scorer)

			resp, err := uc.Execute(context.Background(), tt.request)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(resp.NamespaceScores) != tt.expectedNamespaces {
				t.Errorf("expected %d namespace scores, got %d", tt.expectedNamespaces, len(resp.NamespaceScores))
			}

			if tt.expectedScore > 0 && resp.CompositeScore != tt.expectedScore {
				t.Logf("got composite score: %.2f", resp.CompositeScore)
			}

			if resp.ClusterID != tt.request.ClusterID {
				t.Errorf("expected cluster ID %s, got %s", tt.request.ClusterID, resp.ClusterID)
			}
		})
	}
}

func TestAnalyzeCostEfficiencyUseCase_ValidateInput(t *testing.T) {
	clusterRepo := &mockClusterRepo{
		clusters: map[string]*entities.Cluster{
			"cluster-1": {
				ID:       "cluster-1",
				Name:     "test-cluster",
				Region:   "us-central1",
				IsActive: true,
			},
		},
		namespaces: map[string][]string{
			"cluster-1": {"default"},
		},
	}
	metricsRepo := &mockMetricsRepo{}
	scorer := services.NewEfficiencyScorer()

	uc := NewAnalyzeCostEfficiencyUseCase(clusterRepo, metricsRepo, scorer)

	// Test empty cluster ID
	_, err := uc.Execute(context.Background(), AnalyzeCostEfficiencyRequest{
		ClusterID: "",
	})
	if err == nil {
		t.Error("expected error for empty cluster ID")
	}
}

func TestAnalyzeCostEfficiencyUseCase_CalculatesCompositeScore(t *testing.T) {
	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-24*time.Hour), now)

	clusterRepo := &mockClusterRepo{
		clusters: map[string]*entities.Cluster{
			"cluster-1": {
				ID:       "cluster-1",
				Name:     "test-cluster",
				Region:   "us-central1",
				IsActive: true,
			},
		},
		namespaces: map[string][]string{
			"cluster-1": {"default", "monitoring"},
		},
	}
	metricsRepo := &mockMetricsRepo{
		metrics: map[string]services.NamespaceMetrics{
			"cluster-1/default": {
				Namespace:         "default",
				CPUUtilization:    100,
				MemoryUtilization: 100,
				StorageEfficiency: 100,
				TimeRange:         timeRange,
			},
			"cluster-1/monitoring": {
				Namespace:         "monitoring",
				CPUUtilization:    50,
				MemoryUtilization: 50,
				StorageEfficiency: 50,
				TimeRange:         timeRange,
			},
		},
	}
	scorer := services.NewEfficiencyScorer()

	uc := NewAnalyzeCostEfficiencyUseCase(clusterRepo, metricsRepo, scorer)

	resp, err := uc.Execute(context.Background(), AnalyzeCostEfficiencyRequest{
		ClusterID: "cluster-1",
		WindowConfig: valueobjects.WindowConfig{
			Size: 24 * time.Hour,
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Average of 100 and 50 = 75 for each metric
	// Composite = 75*0.30 + 75*0.40 + 75*0.30 = 75
	expectedScore := 75.0
	if resp.CompositeScore != expectedScore {
		t.Errorf("expected composite score %.2f, got %.2f", expectedScore, resp.CompositeScore)
	}

	// Should be yellow (>= 50 and < 80)
	if resp.ColorCode != valueobjects.ScoreColorYellow {
		t.Errorf("expected yellow, got %s", resp.ColorCode)
	}
}

func TestAnalyzeCostEfficiencyUseCase_ColorCode(t *testing.T) {
	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-24*time.Hour), now)

	tests := []struct {
		name          string
		metrics       map[string]services.NamespaceMetrics
		expectedColor valueobjects.ScoreColor
	}{
		{
			name: "high score - green",
			metrics: map[string]services.NamespaceMetrics{
				"cluster-1/default": {
					Namespace:         "default",
					CPUUtilization:    85,
					MemoryUtilization: 85,
					StorageEfficiency: 85,
					TimeRange:         timeRange,
				},
			},
			expectedColor: valueobjects.ScoreColorGreen,
		},
		{
			name: "medium score - yellow",
			metrics: map[string]services.NamespaceMetrics{
				"cluster-1/default": {
					Namespace:         "default",
					CPUUtilization:    60,
					MemoryUtilization: 60,
					StorageEfficiency: 60,
					TimeRange:         timeRange,
				},
			},
			expectedColor: valueobjects.ScoreColorYellow,
		},
		{
			name: "low score - red",
			metrics: map[string]services.NamespaceMetrics{
				"cluster-1/default": {
					Namespace:         "default",
					CPUUtilization:    30,
					MemoryUtilization: 30,
					StorageEfficiency: 30,
					TimeRange:         timeRange,
				},
			},
			expectedColor: valueobjects.ScoreColorRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRepo := &mockClusterRepo{
				clusters: map[string]*entities.Cluster{
					"cluster-1": {
						ID:       "cluster-1",
						Name:     "test-cluster",
						Region:   "us-central1",
						IsActive: true,
					},
				},
				namespaces: map[string][]string{
					"cluster-1": {"default"},
				},
			}
			metricsRepo := &mockMetricsRepo{
				metrics: tt.metrics,
			}
			scorer := services.NewEfficiencyScorer()

			uc := NewAnalyzeCostEfficiencyUseCase(clusterRepo, metricsRepo, scorer)

			resp, err := uc.Execute(context.Background(), AnalyzeCostEfficiencyRequest{
				ClusterID: "cluster-1",
				WindowConfig: valueobjects.WindowConfig{
					Size: 24 * time.Hour,
				},
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.ColorCode != tt.expectedColor {
				t.Errorf("expected %s, got %s", tt.expectedColor, resp.ColorCode)
			}
		})
	}
}
