package valueobjects

import (
	"testing"
	"time"
)

func TestNewCostEfficiencyScore(t *testing.T) {
	now := time.Now()
	timeRange := NewTimeRange(now.Add(-24*time.Hour), now)

	tests := []struct {
		name          string
		cpu           float64
		memory        float64
		storage       float64
		expectedScore float64
		expectedColor ScoreColor
		expectedConf  ConfidenceLevel
	}{
		{
			name:          "high utilization - green score",
			cpu:           90,
			memory:        85,
			storage:       80,
			expectedScore: 85.0,
			expectedColor: ScoreColorGreen,
			expectedConf:  ConfidenceHigh,
		},
		{
			name:          "medium utilization - yellow score",
			cpu:           60,
			memory:        50,
			storage:       55,
			expectedScore: 54.5,
			expectedColor: ScoreColorYellow,
			expectedConf:  ConfidenceHigh,
		},
		{
			name:          "low utilization - red score",
			cpu:           30,
			memory:        25,
			storage:       20,
			expectedScore: 25.0,
			expectedColor: ScoreColorRed,
			expectedConf:  ConfidenceHigh,
		},
		{
			name:          "boundary at 80 - green",
			cpu:           80,
			memory:        80,
			storage:       80,
			expectedScore: 80.0,
			expectedColor: ScoreColorGreen,
			expectedConf:  ConfidenceHigh,
		},
		{
			name:          "boundary at 50 - yellow",
			cpu:           50,
			memory:        50,
			storage:       50,
			expectedScore: 50.0,
			expectedColor: ScoreColorYellow,
			expectedConf:  ConfidenceHigh,
		},
		{
			name:          "zero values - red with low confidence",
			cpu:           0,
			memory:        0,
			storage:       0,
			expectedScore: 0.0,
			expectedColor: ScoreColorRed,
			expectedConf:  ConfidenceMedium,
		},
		{
			name:          "partial zero values",
			cpu:           50,
			memory:        0,
			storage:       50,
			expectedScore: 30.0,
			expectedColor: ScoreColorRed,
			expectedConf:  ConfidenceMedium,
		},
		{
			name:          "all 100% - perfect score",
			cpu:           100,
			memory:        100,
			storage:       100,
			expectedScore: 100.0,
			expectedColor: ScoreColorGreen,
			expectedConf:  ConfidenceHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := NewCostEfficiencyScore(tt.cpu, tt.memory, tt.storage, timeRange)

			if score.CompositeScore != tt.expectedScore {
				t.Errorf("expected composite score %.2f, got %.2f", tt.expectedScore, score.CompositeScore)
			}
			if score.ColorCode != tt.expectedColor {
				t.Errorf("expected color code %s, got %s", tt.expectedColor, score.ColorCode)
			}
			if score.Confidence != tt.expectedConf {
				t.Errorf("expected confidence %s, got %s", tt.expectedConf, score.Confidence)
			}
			if score.CPUUtilization != tt.cpu {
				t.Errorf("expected cpu utilization %.2f, got %.2f", tt.cpu, score.CPUUtilization)
			}
			if score.MemoryUtilization != tt.memory {
				t.Errorf("expected memory utilization %.2f, got %.2f", tt.memory, score.MemoryUtilization)
			}
			if score.StorageEfficiency != tt.storage {
				t.Errorf("expected storage efficiency %.2f, got %.2f", tt.storage, score.StorageEfficiency)
			}
		})
	}
}

func TestCostEfficiencyScoreCompositeCalculation(t *testing.T) {
	// Test the weighted formula: 30% CPU, 40% Memory, 30% Storage
	now := time.Now()
	timeRange := NewTimeRange(now.Add(-24*time.Hour), now)

	tests := []struct {
		name     string
		cpu      float64
		memory   float64
		storage  float64
		expected float64
	}{
		{
			name:     "equal weights",
			cpu:      50,
			memory:   50,
			storage:  50,
			expected: 50.0, // 50*0.3 + 50*0.4 + 50*0.3 = 50
		},
		{
			name:     "memory heavy",
			cpu:      20,
			memory:   100,
			storage:  20,
			expected: 52.0, // 20*0.3 + 100*0.4 + 20*0.3 = 6 + 40 + 6 = 52
		},
		{
			name:     "cpu heavy",
			cpu:      100,
			memory:   20,
			storage:  20,
			expected: 44.0, // 100*0.3 + 20*0.4 + 20*0.3 = 30 + 8 + 6 = 44
		},
		{
			name:     "storage heavy",
			cpu:      20,
			memory:   20,
			storage:  100,
			expected: 44.0, // 20*0.3 + 20*0.4 + 100*0.3 = 6 + 8 + 30 = 44
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := NewCostEfficiencyScore(tt.cpu, tt.memory, tt.storage, timeRange)

			// Allow small floating point variance
			if abs(score.CompositeScore-tt.expected) > 0.01 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, score.CompositeScore)
			}
		})
	}
}

func abs(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}

func TestNewSavings(t *testing.T) {
	tests := []struct {
		name               string
		monthlySavings     float64
		expectedMonthly    float64
		expectedAnnual     float64
		expectedPercentage float64
		expectedROIMonths  int
	}{
		{
			name:               "positive savings",
			monthlySavings:     100.0,
			expectedMonthly:    100.0,
			expectedAnnual:     1200.0,
			expectedPercentage: 0.0, // Not calculated yet
			expectedROIMonths:  0,
		},
		{
			name:               "zero savings",
			monthlySavings:     0.0,
			expectedMonthly:    0.0,
			expectedAnnual:     0.0,
			expectedPercentage: 0.0,
			expectedROIMonths:  0,
		},
		{
			name:               "negative savings",
			monthlySavings:     -50.0,
			expectedMonthly:    -50.0,
			expectedAnnual:     -600.0,
			expectedPercentage: 0.0,
			expectedROIMonths:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savings := NewSavings(tt.monthlySavings)

			if savings.MonthlySavings != tt.expectedMonthly {
				t.Errorf("expected monthly savings %.2f, got %.2f", tt.expectedMonthly, savings.MonthlySavings)
			}
			if savings.AnnualSavings != tt.expectedAnnual {
				t.Errorf("expected annual savings %.2f, got %.2f", tt.expectedAnnual, savings.AnnualSavings)
			}
			if savings.PercentageReduction != tt.expectedPercentage {
				t.Errorf("expected percentage reduction %.2f, got %.2f", tt.expectedPercentage, savings.PercentageReduction)
			}
			if savings.ROIMonths != tt.expectedROIMonths {
				t.Errorf("expected ROI months %d, got %d", tt.expectedROIMonths, savings.ROIMonths)
			}
		})
	}
}

func TestSavingsCalculateWithCurrentCost(t *testing.T) {
	tests := []struct {
		name               string
		monthlySavings     float64
		currentCost        float64
		expectedPercentage float64
	}{
		{
			name:               "normal savings calculation",
			monthlySavings:     100.0,
			currentCost:        500.0,
			expectedPercentage: 20.0, // 100/500 * 100 = 20%
		},
		{
			name:               "zero current cost",
			monthlySavings:     100.0,
			currentCost:        0.0,
			expectedPercentage: 0.0, // Division by zero
		},
		{
			name:               "negative current cost",
			monthlySavings:     100.0,
			currentCost:        -100.0,
			expectedPercentage: 0.0, // Division by zero (negative)
		},
		{
			name:               "100% savings",
			monthlySavings:     500.0,
			currentCost:        500.0,
			expectedPercentage: 100.0,
		},
		{
			name:               "small savings",
			monthlySavings:     10.0,
			currentCost:        1000.0,
			expectedPercentage: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savings := NewSavings(tt.monthlySavings)
			savings.CalculateWithCurrentCost(tt.currentCost)

			if savings.PercentageReduction != tt.expectedPercentage {
				t.Errorf("expected percentage %.2f, got %.2f", tt.expectedPercentage, savings.PercentageReduction)
			}
		})
	}
}

func TestNewPricingInfo(t *testing.T) {
	tests := []struct {
		name                   string
		machineType            string
		machineFamily          string
		region                 string
		onDemand               float64
		spot                   float64
		expectedOnDemand       float64
		expectedSpot           float64
		expectedSpotSavings    float64
		expectedPersistentCost float64
	}{
		{
			name:                   "standard pricing",
			machineType:            "e2-standard-4",
			machineFamily:          "e2",
			region:                 "us-central1",
			onDemand:               0.20,
			spot:                   0.06,
			expectedOnDemand:       0.20,
			expectedSpot:           0.06,
			expectedSpotSavings:    70.0, // (0.20-0.06)/0.20 * 100 = 70%
			expectedPersistentCost: 0.0,
		},
		{
			name:                   "zero on demand",
			machineType:            "n2-standard-4",
			machineFamily:          "n2",
			region:                 "us-central1",
			onDemand:               0.0,
			spot:                   0.0,
			expectedOnDemand:       0.0,
			expectedSpot:           0.0,
			expectedSpotSavings:    0.0,
			expectedPersistentCost: 0.0,
		},
		{
			name:                   "90% savings",
			machineType:            "c3-standard-4",
			machineFamily:          "c3",
			region:                 "us-east1",
			onDemand:               1.0,
			spot:                   0.1,
			expectedOnDemand:       1.0,
			expectedSpot:           0.1,
			expectedSpotSavings:    90.0,
			expectedPersistentCost: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing := NewPricingInfo(tt.machineType, tt.machineFamily, tt.region, tt.onDemand, tt.spot)

			if pricing.MachineType != tt.machineType {
				t.Errorf("expected machine type %s, got %s", tt.machineType, pricing.MachineType)
			}
			if pricing.MachineFamily != tt.machineFamily {
				t.Errorf("expected machine family %s, got %s", tt.machineFamily, pricing.MachineFamily)
			}
			if pricing.Region != tt.region {
				t.Errorf("expected region %s, got %s", tt.region, pricing.Region)
			}
			if pricing.OnDemandHourly != tt.expectedOnDemand {
				t.Errorf("expected on demand %.4f, got %.4f", tt.expectedOnDemand, pricing.OnDemandHourly)
			}
			if pricing.SpotHourly != tt.expectedSpot {
				t.Errorf("expected spot %.4f, got %.4f", tt.expectedSpot, pricing.SpotHourly)
			}
			if pricing.SpotSavingsPercent != tt.expectedSpotSavings {
				t.Errorf("expected spot savings %.2f, got %.2f", tt.expectedSpotSavings, pricing.SpotSavingsPercent)
			}
		})
	}
}

func TestPricingInfoGetMonthlyCost(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		pricingType     PricingType
		hoursPerMonth   float64
		onDemandHourly  float64
		spotHourly      float64
		committedHourly float64
		expectedCost    float64
	}{
		{
			name:            "on demand pricing",
			pricingType:     PricingTypeOnDemand,
			hoursPerMonth:   730,
			onDemandHourly:  0.20,
			spotHourly:      0.06,
			committedHourly: 0.14,
			expectedCost:    146.0, // 0.20 * 730
		},
		{
			name:            "spot pricing",
			pricingType:     PricingTypeSpot,
			hoursPerMonth:   730,
			onDemandHourly:  0.20,
			spotHourly:      0.06,
			committedHourly: 0.14,
			expectedCost:    43.8, // 0.06 * 730
		},
		{
			name:            "committed pricing",
			pricingType:     PricingTypeCommitted,
			hoursPerMonth:   730,
			onDemandHourly:  0.20,
			spotHourly:      0.06,
			committedHourly: 0.14,
			expectedCost:    102.2, // 0.14 * 730
		},
		{
			name:            "zero hours",
			pricingType:     PricingTypeOnDemand,
			hoursPerMonth:   0,
			onDemandHourly:  0.20,
			spotHourly:      0.06,
			committedHourly: 0.14,
			expectedCost:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing := NewPricingInfo("e2-standard-4", "e2", "us-central1", tt.onDemandHourly, tt.spotHourly)
			pricing.PricingType = tt.pricingType
			pricing.CommittedHourly = tt.committedHourly

			cost := pricing.GetMonthlyCost(tt.hoursPerMonth)

			if cost != tt.expectedCost {
				t.Errorf("expected cost %.2f, got %.2f", tt.expectedCost, cost)
			}
		})
	}

	// Test with time object (default values)
	_ = now
}

func TestNewTimeRange(t *testing.T) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	tr := NewTimeRange(start, end)

	if !tr.StartTime.Equal(start) {
		t.Errorf("expected start time %v, got %v", start, tr.StartTime)
	}
	if !tr.EndTime.Equal(end) {
		t.Errorf("expected end time %v, got %v", end, tr.EndTime)
	}
	// Allow for small floating point variance in duration
	if tr.Duration < 23*time.Hour || tr.Duration > 25*time.Hour {
		t.Errorf("expected duration approximately 24h, got %v", tr.Duration)
	}
}

func TestRecommendationAccept(t *testing.T) {
	now := time.Now()
	rec := &Recommendation{
		ID:        "test-rec",
		Status:    RecommendationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	rec.Accept()

	if rec.Status != RecommendationStatusAccepted {
		t.Errorf("expected status %s, got %s", RecommendationStatusAccepted, rec.Status)
	}
	if rec.UpdatedAt.Before(now) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestRecommendationReject(t *testing.T) {
	now := time.Now()
	rec := &Recommendation{
		ID:        "test-rec",
		Status:    RecommendationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	rec.Reject()

	if rec.Status != RecommendationStatusRejected {
		t.Errorf("expected status %s, got %s", RecommendationStatusRejected, rec.Status)
	}
	if rec.UpdatedAt.Before(now) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestRecommendationSetSavings(t *testing.T) {
	rec := &Recommendation{
		ID:     "test-rec",
		Status: RecommendationStatusPending,
	}

	savings := NewSavings(100.0)
	rec.SetSavings(savings)

	if rec.Savings != savings {
		t.Error("expected savings to be set")
	}
	if rec.Savings.MonthlySavings != 100.0 {
		t.Errorf("expected monthly savings %.2f, got %.2f", 100.0, rec.Savings.MonthlySavings)
	}
}

func TestDefaultWindowConfigs(t *testing.T) {
	configs := DefaultWindowConfigs()

	if len(configs) != 4 {
		t.Errorf("expected 4 window configs, got %d", len(configs))
	}

	// Check 1h config
	if configs["1h"].Size != time.Hour {
		t.Errorf("expected 1h size %v, got %v", time.Hour, configs["1h"].Size)
	}
	if configs["1h"].Percent != 95 {
		t.Errorf("expected 1h percent 95, got %d", configs["1h"].Percent)
	}

	// Check 24h config
	if configs["24h"].Size != 24*time.Hour {
		t.Errorf("expected 24h size %v, got %v", 24*time.Hour, configs["24h"].Size)
	}

	// Check 7d config
	if configs["7d"].Size != 7*24*time.Hour {
		t.Errorf("expected 7d size %v, got %v", 7*24*time.Hour, configs["7d"].Size)
	}
}
