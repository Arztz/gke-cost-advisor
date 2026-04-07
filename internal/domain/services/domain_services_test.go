package services

import (
	"math"
	"testing"
	"time"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/valueobjects"
)

// Helper function to compare floats with tolerance
func floatEquals(expected, actual, tolerance float64) bool {
	return math.Abs(expected-actual) <= tolerance
}

func TestCostCalculatorCalculateCurrentCost(t *testing.T) {
	calculator := NewCostCalculator()

	tests := []struct {
		name         string
		resources    entities.ResourceRequirements
		pricing      *valueobjects.PricingInfo
		expectedCost float64
	}{
		{
			name: "with CPU and memory requests",
			resources: entities.ResourceRequirements{
				CPURequest:    &entities.ResourceQuantity{MilliCPU: 2000},                // 2 cores
				MemoryRequest: &entities.ResourceQuantity{Bytes: 4 * 1024 * 1024 * 1024}, // 4GB
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly:     0.20,
				PersistentDiskCost: 0.02, // per GB per month
			},
			// CPU: 0.20 * 2 = 0.40 per hour
			// Memory: 4 * 0.02 = 0.08 per hour (treated as hourly in calculation)
			// Total per hour: 0.40 + 0.08 = 0.48
			// Monthly (730 hours): 0.48 * 730 = 350.4
			expectedCost: 350.4,
		},
		{
			name: "with CPU only",
			resources: entities.ResourceRequirements{
				CPURequest: &entities.ResourceQuantity{MilliCPU: 1000}, // 1 core
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly:     0.20,
				PersistentDiskCost: 0.02,
			},
			// CPU: 0.20 * 1 = 0.20 per hour
			// Monthly: 0.20 * 730 = 146.0
			expectedCost: 146.0,
		},
		{
			name: "with memory only",
			resources: entities.ResourceRequirements{
				MemoryRequest: &entities.ResourceQuantity{Bytes: 2 * 1024 * 1024 * 1024}, // 2GB
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly:     0.20,
				PersistentDiskCost: 0.02,
			},
			// Memory: 2 * 0.02 = 0.04 per hour
			// Monthly: 0.04 * 730 = 29.2
			expectedCost: 29.2,
		},
		{
			name:      "with no resources",
			resources: entities.ResourceRequirements{},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly:     0.20,
				PersistentDiskCost: 0.02,
			},
			expectedCost: 0.0,
		},
		{
			name: "with zero pricing",
			resources: entities.ResourceRequirements{
				CPURequest:    &entities.ResourceQuantity{MilliCPU: 1000},
				MemoryRequest: &entities.ResourceQuantity{Bytes: 1024 * 1024 * 1024},
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly:     0.0,
				PersistentDiskCost: 0.0,
			},
			expectedCost: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculator.CalculateCurrentCost(tt.resources, tt.pricing)
			if !floatEquals(tt.expectedCost, cost, 0.01) {
				t.Errorf("expected cost %.2f, got %.2f", tt.expectedCost, cost)
			}
		})
	}
}

func TestCostCalculatorCalculateProjectedCost(t *testing.T) {
	calculator := NewCostCalculator()

	tests := []struct {
		name         string
		resources    entities.ResourceRequirements
		pricing      *valueobjects.PricingInfo
		strategy     OptimizationStrategy
		expectedCost float64
	}{
		{
			name: "on demand strategy",
			resources: entities.ResourceRequirements{
				CPURequest: &entities.ResourceQuantity{MilliCPU: 1000},
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly: 0.20,
			},
			strategy:     StrategyOnDemand,
			expectedCost: 146.0, // 0.20 * 1 * 730
		},
		{
			name: "spot strategy",
			resources: entities.ResourceRequirements{
				CPURequest: &entities.ResourceQuantity{MilliCPU: 1000},
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly: 0.20,
				SpotHourly:     0.06,
			},
			strategy:     StrategySpot,
			expectedCost: 43.8, // 0.06 * 1 * 730
		},
		{
			name: "committed strategy",
			resources: entities.ResourceRequirements{
				CPURequest: &entities.ResourceQuantity{MilliCPU: 1000},
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly:  0.20,
				CommittedHourly: 0.14,
			},
			strategy:     StrategyCommitted,
			expectedCost: 102.2, // 0.14 * 1 * 730
		},
		{
			name: "default strategy (on demand)",
			resources: entities.ResourceRequirements{
				CPURequest: &entities.ResourceQuantity{MilliCPU: 1000},
			},
			pricing: &valueobjects.PricingInfo{
				OnDemandHourly: 0.20,
			},
			strategy:     StrategyRightSize, // falls back to on demand
			expectedCost: 146.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculator.CalculateProjectedCost(tt.resources, tt.pricing, tt.strategy)
			if cost != tt.expectedCost {
				t.Errorf("expected cost %.2f, got %.2f", tt.expectedCost, cost)
			}
		})
	}
}

func TestCostCalculatorCalculateSpotSavings(t *testing.T) {
	calculator := NewCostCalculator()

	savings := calculator.CalculateSpotSavings(100.0, 30.0)

	if savings.MonthlySavings != 70.0 {
		t.Errorf("expected monthly savings 70.0, got %.2f", savings.MonthlySavings)
	}
	if savings.AnnualSavings != 840.0 {
		t.Errorf("expected annual savings 840.0, got %.2f", savings.AnnualSavings)
	}

	// Test with current cost
	savings.CalculateWithCurrentCost(100.0)
	if savings.PercentageReduction != 70.0 {
		t.Errorf("expected percentage reduction 70.0, got %.2f", savings.PercentageReduction)
	}
}

func TestCostCalculatorCalculateAnnualProjection(t *testing.T) {
	calculator := NewCostCalculator()

	annual := calculator.CalculateAnnualProjection(100.0)
	if annual != 1200.0 {
		t.Errorf("expected 1200.0, got %.2f", annual)
	}

	annual = calculator.CalculateAnnualProjection(0.0)
	if annual != 0.0 {
		t.Errorf("expected 0.0, got %.2f", annual)
	}
}

func TestCostCalculatorCalculateNodePoolCost(t *testing.T) {
	calculator := NewCostCalculator()

	pricing := &valueobjects.PricingInfo{
		OnDemandHourly: 0.20,
	}

	cost := calculator.CalculateNodePoolCost("e2-standard-4", 3, pricing)
	// 0.20 * 3 * 730 = 438.0
	if !floatEquals(438.0, cost, 0.01) {
		t.Errorf("expected 438.0, got %.2f", cost)
	}

	cost = calculator.CalculateNodePoolCost("n2-standard-4", 0, pricing)
	if cost != 0.0 {
		t.Errorf("expected 0.0 for zero nodes, got %.2f", cost)
	}
}

func TestEfficiencyScorerCalculateNamespaceScore(t *testing.T) {
	scorer := NewEfficiencyScorer()
	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-24*time.Hour), now)

	tests := []struct {
		name          string
		metrics       NamespaceMetrics
		expectedScore float64
		expectedColor valueobjects.ScoreColor
	}{
		{
			name: "high utilization",
			metrics: NamespaceMetrics{
				Namespace:         "default",
				CPUUtilization:    90,
				MemoryUtilization: 85,
				StorageEfficiency: 80,
				TimeRange:         timeRange,
			},
			expectedScore: 85.0,
			expectedColor: valueobjects.ScoreColorGreen,
		},
		{
			name: "low utilization",
			metrics: NamespaceMetrics{
				Namespace:         "default",
				CPUUtilization:    20,
				MemoryUtilization: 30,
				StorageEfficiency: 25,
				TimeRange:         timeRange,
			},
			expectedScore: 25.5,
			expectedColor: valueobjects.ScoreColorRed,
		},
		{
			name: "zero utilization",
			metrics: NamespaceMetrics{
				Namespace:         "default",
				CPUUtilization:    0,
				MemoryUtilization: 0,
				StorageEfficiency: 0,
				TimeRange:         timeRange,
			},
			expectedScore: 0.0,
			expectedColor: valueobjects.ScoreColorRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.CalculateNamespaceScore(tt.metrics)

			if score.CompositeScore != tt.expectedScore {
				t.Errorf("expected score %.2f, got %.2f", tt.expectedScore, score.CompositeScore)
			}
			if score.ColorCode != tt.expectedColor {
				t.Errorf("expected color %s, got %s", tt.expectedColor, score.ColorCode)
			}
			if score.Confidence != valueobjects.ConfidenceHigh && tt.metrics.CPUUtilization > 0 {
				t.Errorf("expected high confidence, got %s", score.Confidence)
			}
		})
	}
}

func TestEfficiencyScorerCalculateClusterScore(t *testing.T) {
	scorer := NewEfficiencyScorer()
	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-24*time.Hour), now)

	cluster := &entities.Cluster{
		ID:   "cluster-1",
		Name: "test-cluster",
	}

	tests := []struct {
		name          string
		metricsList   []NamespaceMetrics
		expectedScore float64
		expectedColor valueobjects.ScoreColor
	}{
		{
			name: "single namespace",
			metricsList: []NamespaceMetrics{
				{
					CPUUtilization:    80,
					MemoryUtilization: 80,
					StorageEfficiency: 80,
					TimeRange:         timeRange,
				},
			},
			expectedScore: 80.0,
			expectedColor: valueobjects.ScoreColorGreen,
		},
		{
			name: "multiple namespaces average",
			metricsList: []NamespaceMetrics{
				{
					CPUUtilization:    100,
					MemoryUtilization: 100,
					StorageEfficiency: 100,
					TimeRange:         timeRange,
				},
				{
					CPUUtilization:    50,
					MemoryUtilization: 50,
					StorageEfficiency: 50,
					TimeRange:         timeRange,
				},
			},
			expectedScore: 75.0,
			expectedColor: valueobjects.ScoreColorYellow,
		},
		{
			name:          "empty metrics list",
			metricsList:   []NamespaceMetrics{},
			expectedScore: 0.0,
			expectedColor: valueobjects.ScoreColorRed,
		},
		{
			name:          "nil metrics list",
			metricsList:   nil,
			expectedScore: 0.0,
			expectedColor: valueobjects.ScoreColorRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.CalculateClusterScore(cluster, tt.metricsList)

			if score.CompositeScore != tt.expectedScore {
				t.Errorf("expected score %.2f, got %.2f", tt.expectedScore, score.CompositeScore)
			}
			if score.ColorCode != tt.expectedColor {
				t.Errorf("expected color %s, got %s", tt.expectedColor, score.ColorCode)
			}
		})
	}
}

func TestEfficiencyScorerCalculateWorkloadScore(t *testing.T) {
	scorer := NewEfficiencyScorer()
	now := time.Now()
	timeRange := valueobjects.NewTimeRange(now.Add(-24*time.Hour), now)

	workload := &entities.Workload{
		ID:        "workload-1",
		Name:      "test-workload",
		Namespace: "default",
	}

	metrics := NamespaceMetrics{
		Namespace:         "default",
		CPUUtilization:    75,
		MemoryUtilization: 80,
		StorageEfficiency: 70,
		TimeRange:         timeRange,
	}

	score := scorer.CalculateWorkloadScore(workload, metrics)

	// Should return same as CalculateNamespaceScore
	expectedScore := (75*0.30 + 80*0.40 + 70*0.30)
	if score.CompositeScore != expectedScore {
		t.Errorf("expected score %.2f, got %.2f", expectedScore, score.CompositeScore)
	}
	if score.ColorCode != valueobjects.ScoreColorYellow {
		t.Errorf("expected yellow, got %s", score.ColorCode)
	}
}

func TestRightSizerGenerateCPURequestRecommendation(t *testing.T) {
	rightSizer := NewRightSizer()

	tests := []struct {
		name             string
		percentiles      valueobjects.PercentileData
		latencyProfile   valueobjects.LatencyProfile
		expectedHeadroom float64
		expectedConf     valueobjects.ConfidenceLevel
	}{
		{
			name: "default latency profile uses P95",
			percentiles: valueobjects.PercentileData{
				P50CPU: 0.5,
				P90CPU: 1.0,
				P95CPU: 2.0,
				P99CPU: 3.0,
			},
			latencyProfile:   valueobjects.LatencyProfileMedium,
			expectedHeadroom: 20.0, // 20% default headroom
			expectedConf:     valueobjects.ConfidenceMedium,
		},
		{
			name: "low latency profile uses P90",
			percentiles: valueobjects.PercentileData{
				P50CPU: 0.5,
				P90CPU: 1.0,
				P95CPU: 2.0,
				P99CPU: 3.0,
			},
			latencyProfile:   valueobjects.LatencyProfileLow,
			expectedHeadroom: 20.0,
			expectedConf:     valueobjects.ConfidenceLow, // variability = 0.5, which triggers low confidence
		},
		{
			name: "high confidence with low variability",
			percentiles: valueobjects.PercentileData{
				P50CPU: 1.0,
				P90CPU: 1.8,
				P95CPU: 2.0,
				P99CPU: 2.2, // 10% variability (2.2-2.0)/2.0
			},
			latencyProfile:   valueobjects.LatencyProfileMedium,
			expectedHeadroom: 20.0,
			expectedConf:     valueobjects.ConfidenceHigh,
		},
		{
			name: "low confidence with high variability",
			percentiles: valueobjects.PercentileData{
				P50CPU: 1.0,
				P90CPU: 1.5,
				P95CPU: 2.0,
				P99CPU: 4.0, // 100% variability (4.0-2.0)/2.0 = 1.0 > 0.5
			},
			latencyProfile:   valueobjects.LatencyProfileMedium,
			expectedHeadroom: 20.0,
			expectedConf:     valueobjects.ConfidenceLow,
		},
		{
			name: "zero percentiles",
			percentiles: valueobjects.PercentileData{
				P50CPU: 0,
				P90CPU: 0,
				P95CPU: 0,
				P99CPU: 0,
			},
			latencyProfile:   valueobjects.LatencyProfileMedium,
			expectedHeadroom: 20.0,
			expectedConf:     valueobjects.ConfidenceMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := rightSizer.GenerateCPURequestRecommendation(tt.percentiles, tt.latencyProfile)

			if rec.ResourceType != "cpu_request" {
				t.Errorf("expected resource type cpu_request, got %s", rec.ResourceType)
			}
			if rec.HeadroomPercent != tt.expectedHeadroom {
				t.Errorf("expected headroom %.2f, got %.2f", tt.expectedHeadroom, rec.HeadroomPercent)
			}
			if rec.Confidence != tt.expectedConf {
				t.Errorf("expected confidence %s, got %s", tt.expectedConf, rec.Confidence)
			}
		})
	}
}

func TestRightSizerGenerateMemoryRequestRecommendation(t *testing.T) {
	rightSizer := NewRightSizer()

	tests := []struct {
		name             string
		percentiles      valueobjects.PercentileData
		latencyProfile   valueobjects.LatencyProfile
		expectedHeadroom float64
		expectedConf     valueobjects.ConfidenceLevel
	}{
		{
			name: "default latency profile uses P95",
			percentiles: valueobjects.PercentileData{
				P50Memory: 1 * 1024 * 1024 * 1024,
				P90Memory: 2 * 1024 * 1024 * 1024,
				P95Memory: 4 * 1024 * 1024 * 1024,
				P99Memory: 6 * 1024 * 1024 * 1024,
			},
			latencyProfile:   valueobjects.LatencyProfileMedium,
			expectedHeadroom: 30.0, // 30% default headroom for memory
			expectedConf:     valueobjects.ConfidenceMedium,
		},
		{
			name: "low latency profile uses P90",
			percentiles: valueobjects.PercentileData{
				P50Memory: 1 * 1024 * 1024 * 1024,
				P90Memory: 2 * 1024 * 1024 * 1024,
				P95Memory: 4 * 1024 * 1024 * 1024,
				P99Memory: 6 * 1024 * 1024 * 1024,
			},
			latencyProfile:   valueobjects.LatencyProfileLow,
			expectedHeadroom: 30.0,
			expectedConf:     valueobjects.ConfidenceMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := rightSizer.GenerateMemoryRequestRecommendation(tt.percentiles, tt.latencyProfile)

			if rec.ResourceType != "memory_request" {
				t.Errorf("expected resource type memory_request, got %s", rec.ResourceType)
			}
			if rec.HeadroomPercent != tt.expectedHeadroom {
				t.Errorf("expected headroom %.2f, got %.2f", tt.expectedHeadroom, rec.HeadroomPercent)
			}
		})
	}
}

func TestRightSizerGenerateCPULimitRecommendation(t *testing.T) {
	rightSizer := NewRightSizer()

	percentiles := valueobjects.PercentileData{
		P99CPU: 2.0,
	}

	rec := rightSizer.GenerateCPULimitRecommendation(percentiles, valueobjects.LatencyProfileMedium)

	if rec.ResourceType != "cpu_limit" {
		t.Errorf("expected resource type cpu_limit, got %s", rec.ResourceType)
	}
	if rec.HeadroomPercent != 10.0 {
		t.Errorf("expected headroom 10.0, got %.2f", rec.HeadroomPercent)
	}
	// 2.0 * 1.10 = 2.2
	if rec.RecommendedValue != 2.2 {
		t.Errorf("expected recommended value 2.2, got %.2f", rec.RecommendedValue)
	}
}

func TestRightSizerGenerateMemoryLimitRecommendation(t *testing.T) {
	rightSizer := NewRightSizer()

	percentiles := valueobjects.PercentileData{
		P99Memory: 4 * 1024 * 1024 * 1024,
	}

	rec := rightSizer.GenerateMemoryLimitRecommendation(percentiles, valueobjects.LatencyProfileMedium)

	if rec.ResourceType != "memory_limit" {
		t.Errorf("expected resource type memory_limit, got %s", rec.ResourceType)
	}
	if rec.HeadroomPercent != 20.0 {
		t.Errorf("expected headroom 20.0, got %.2f", rec.HeadroomPercent)
	}
	// 4GB * 1.20 = 4.8GB
	expected := 4.8 * 1024 * 1024 * 1024
	if rec.RecommendedValue != expected {
		t.Errorf("expected recommended value %.2f, got %.2f", expected, rec.RecommendedValue)
	}
}

func TestRightSizerCalculateWasteGap(t *testing.T) {
	rightSizer := NewRightSizer()

	tests := []struct {
		name             string
		container        *entities.Container
		percentiles      valueobjects.PercentileData
		expectedCPUWaste float64
		expectedMemWaste float64
	}{
		{
			name: "high CPU waste",
			container: &entities.Container{
				ID:   "container-1",
				Name: "test",
				Resources: entities.ResourceRequirements{
					CPURequest:    &entities.ResourceQuantity{MilliCPU: 4000}, // 4 cores
					MemoryRequest: &entities.ResourceQuantity{Bytes: 0},
				},
			},
			percentiles: valueobjects.PercentileData{
				P95CPU: 1.0, // Only using 1 core
			},
			expectedCPUWaste: 75.0, // (4-1)/4 * 100 = 75%
			expectedMemWaste: 0.0,
		},
		{
			name: "high memory waste",
			container: &entities.Container{
				ID:   "container-1",
				Name: "test",
				Resources: entities.ResourceRequirements{
					CPURequest:    &entities.ResourceQuantity{MilliCPU: 0},
					MemoryRequest: &entities.ResourceQuantity{Bytes: 8 * 1024 * 1024 * 1024}, // 8GB
				},
			},
			percentiles: valueobjects.PercentileData{
				P95Memory: 2 * 1024 * 1024 * 1024, // Only using 2GB
			},
			expectedCPUWaste: 0.0,
			expectedMemWaste: 75.0, // (8-2)/8 * 100 = 75%
		},
		{
			name: "no waste - using all requested",
			container: &entities.Container{
				ID:   "container-1",
				Name: "test",
				Resources: entities.ResourceRequirements{
					CPURequest:    &entities.ResourceQuantity{MilliCPU: 2000},
					MemoryRequest: &entities.ResourceQuantity{Bytes: 2 * 1024 * 1024 * 1024},
				},
			},
			percentiles: valueobjects.PercentileData{
				P95CPU:    2.0,
				P95Memory: 2 * 1024 * 1024 * 1024,
			},
			expectedCPUWaste: 0.0,
			expectedMemWaste: 0.0,
		},
		{
			name: "no waste - utilization exceeds request",
			container: &entities.Container{
				ID:   "container-1",
				Name: "test",
				Resources: entities.ResourceRequirements{
					CPURequest:    &entities.ResourceQuantity{MilliCPU: 1000},
					MemoryRequest: &entities.ResourceQuantity{Bytes: 1 * 1024 * 1024 * 1024},
				},
			},
			percentiles: valueobjects.PercentileData{
				P95CPU:    3.0,                    // Using more than requested
				P95Memory: 4 * 1024 * 1024 * 1024, // Using more than requested
			},
			expectedCPUWaste: 0.0,
			expectedMemWaste: 0.0,
		},
		{
			name: "zero request - no waste",
			container: &entities.Container{
				ID:   "container-1",
				Name: "test",
				Resources: entities.ResourceRequirements{
					CPURequest:    &entities.ResourceQuantity{MilliCPU: 0},
					MemoryRequest: &entities.ResourceQuantity{Bytes: 0},
				},
			},
			percentiles: valueobjects.PercentileData{
				P95CPU:    1.0,
				P95Memory: 1 * 1024 * 1024 * 1024,
			},
			expectedCPUWaste: 0.0,
			expectedMemWaste: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpuWaste, memWaste := rightSizer.CalculateWasteGap(tt.container, tt.percentiles)

			if cpuWaste != tt.expectedCPUWaste {
				t.Errorf("expected CPU waste %.2f, got %.2f", tt.expectedCPUWaste, cpuWaste)
			}
			if memWaste != tt.expectedMemWaste {
				t.Errorf("expected memory waste %.2f, got %.2f", tt.expectedMemWaste, memWaste)
			}
		})
	}
}

func TestRightSizerGenerateBatchRecommendations(t *testing.T) {
	rightSizer := NewRightSizer()

	containers := []*entities.Container{}
	percentiles := map[string]valueobjects.PercentileData{}

	recs, err := rightSizer.GenerateBatchRecommendations(containers, percentiles)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected empty recommendations, got %d", len(recs))
	}
}
