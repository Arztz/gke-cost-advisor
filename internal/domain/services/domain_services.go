package services

import (
	"math"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/valueobjects"
)

// costCalculator implements the CostCalculator interface
type costCalculator struct {
	hoursPerMonth float64
}

// NewCostCalculator creates a new cost calculator service
func NewCostCalculator() CostCalculator {
	return &costCalculator{
		hoursPerMonth: 730, // Average hours per month
	}
}

// CalculateCurrentCost calculates current cost based on resources and pricing
func (c *costCalculator) CalculateCurrentCost(resources entities.ResourceRequirements, pricing *valueobjects.PricingInfo) float64 {
	var cpuCost, memoryCost float64

	if resources.CPURequest != nil {
		cpuCost = pricing.OnDemandHourly * float64(resources.CPURequest.MilliCPU) / 1000
	}
	if resources.MemoryRequest != nil {
		memoryCost = float64(resources.MemoryRequest.Bytes) / (1024 * 1024 * 1024) * pricing.PersistentDiskCost
	}

	return (cpuCost + memoryCost) * c.hoursPerMonth
}

// CalculateProjectedCost calculates projected cost with optimization strategy
func (c *costCalculator) CalculateProjectedCost(resources entities.ResourceRequirements, pricing *valueobjects.PricingInfo, strategy OptimizationStrategy) float64 {
	var hourlyRate float64

	switch strategy {
	case StrategySpot:
		hourlyRate = pricing.SpotHourly
	case StrategyCommitted:
		hourlyRate = pricing.CommittedHourly
	default:
		hourlyRate = pricing.OnDemandHourly
	}

	var cpuCost, memoryCost float64
	if resources.CPURequest != nil {
		cpuCost = hourlyRate * float64(resources.CPURequest.MilliCPU) / 1000
	}
	if resources.MemoryRequest != nil {
		memoryCost = float64(resources.MemoryRequest.Bytes) / (1024 * 1024 * 1024) * pricing.PersistentDiskCost
	}

	return (cpuCost + memoryCost) * c.hoursPerMonth
}

// CalculateSpotSavings calculates savings from using Spot VMs
func (c *costCalculator) CalculateSpotSavings(onDemandCost float64, spotCost float64) *valueobjects.Savings {
	monthlySavings := onDemandCost - spotCost
	savings := valueobjects.NewSavings(monthlySavings)
	savings.CalculateWithCurrentCost(onDemandCost)
	return savings
}

// CalculateAnnualProjection calculates annual projection from monthly cost
func (c *costCalculator) CalculateAnnualProjection(monthlyCost float64) float64 {
	return monthlyCost * 12
}

// CalculateNodePoolCost calculates the monthly cost for a node pool
func (c *costCalculator) CalculateNodePoolCost(machineType string, nodeCount int, pricing *valueobjects.PricingInfo) float64 {
	hourlyCost := pricing.OnDemandHourly * float64(nodeCount)
	return hourlyCost * c.hoursPerMonth
}

// efficiencyScorer implements the EfficiencyScorer interface
type efficiencyScorer struct{}

// NewEfficiencyScorer creates a new efficiency scorer service
func NewEfficiencyScorer() EfficiencyScorer {
	return &efficiencyScorer{}
}

// CalculateNamespaceScore calculates cost efficiency score for a namespace
func (s *efficiencyScorer) CalculateNamespaceScore(metrics NamespaceMetrics) valueobjects.CostEfficiencyScore {
	score := valueobjects.NewCostEfficiencyScore(
		metrics.CPUUtilization,
		metrics.MemoryUtilization,
		metrics.StorageEfficiency,
		metrics.TimeRange,
	)

	return score
}

// CalculateWorkloadScore calculates score for a workload
func (s *efficiencyScorer) CalculateWorkloadScore(workload *entities.Workload, metrics NamespaceMetrics) valueobjects.CostEfficiencyScore {
	return s.CalculateNamespaceScore(metrics)
}

// CalculateClusterScore calculates score for a cluster
func (s *efficiencyScorer) CalculateClusterScore(cluster *entities.Cluster, metricsList []NamespaceMetrics) valueobjects.CostEfficiencyScore {
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
	avgCPU := totalCPU / count
	avgMemory := totalMemory / count
	avgStorage := totalStorage / count

	return valueobjects.NewCostEfficiencyScore(avgCPU, avgMemory, avgStorage, metricsList[0].TimeRange)
}

// rightSizer implements the RightSizer interface
type rightSizer struct {
	defaultCPUHeadroom    float64
	defaultMemoryHeadroom float64
}

// NewRightSizer creates a new right-sizing service
func NewRightSizer() RightSizer {
	return &rightSizer{
		defaultCPUHeadroom:    0.20,
		defaultMemoryHeadroom: 0.30,
	}
}

// GenerateCPURequestRecommendation generates CPU request recommendation
func (r *rightSizer) GenerateCPURequestRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation {
	// Use P95 for latency-sensitive, P90 for others
	p95CPU := utilizationPercentiles.P95CPU
	if latencyProfile == valueobjects.LatencyProfileLow {
		p95CPU = utilizationPercentiles.P90CPU
	}

	recommended := p95CPU * (1 + r.defaultCPUHeadroom)

	// Determine confidence based on data quality
	confidence := valueobjects.ConfidenceMedium
	if p95CPU > 0 && utilizationPercentiles.P99CPU > 0 {
		variability := (utilizationPercentiles.P99CPU - p95CPU) / p95CPU
		if variability < 0.2 {
			confidence = valueobjects.ConfidenceHigh
		} else if variability > 0.5 {
			confidence = valueobjects.ConfidenceLow
		}
	}

	return valueobjects.ResourceRecommendation{
		ResourceType:     "cpu_request",
		CurrentValue:     p95CPU,
		RecommendedValue: recommended,
		HeadroomPercent:  r.defaultCPUHeadroom * 100,
		Confidence:       confidence,
	}
}

// GenerateMemoryRequestRecommendation generates memory request recommendation
func (r *rightSizer) GenerateMemoryRequestRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation {
	p95Memory := utilizationPercentiles.P95Memory
	if latencyProfile == valueobjects.LatencyProfileLow {
		p95Memory = utilizationPercentiles.P90Memory
	}

	recommended := p95Memory * (1 + r.defaultMemoryHeadroom)

	confidence := valueobjects.ConfidenceMedium
	if p95Memory > 0 && utilizationPercentiles.P99Memory > 0 {
		variability := (utilizationPercentiles.P99Memory - p95Memory) / p95Memory
		if variability < 0.2 {
			confidence = valueobjects.ConfidenceHigh
		} else if variability > 0.5 {
			confidence = valueobjects.ConfidenceLow
		}
	}

	return valueobjects.ResourceRecommendation{
		ResourceType:     "memory_request",
		CurrentValue:     p95Memory,
		RecommendedValue: recommended,
		HeadroomPercent:  r.defaultMemoryHeadroom * 100,
		Confidence:       confidence,
	}
}

// GenerateCPULimitRecommendation generates CPU limit recommendation
func (r *rightSizer) GenerateCPULimitRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation {
	p99CPU := utilizationPercentiles.P99CPU
	recommended := p99CPU * 1.10 // 10% headroom

	return valueobjects.ResourceRecommendation{
		ResourceType:     "cpu_limit",
		CurrentValue:     p99CPU,
		RecommendedValue: recommended,
		HeadroomPercent:  10,
		Confidence:       valueobjects.ConfidenceMedium,
	}
}

// GenerateMemoryLimitRecommendation generates memory limit recommendation
func (r *rightSizer) GenerateMemoryLimitRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation {
	p99Memory := utilizationPercentiles.P99Memory
	recommended := p99Memory * 1.20 // 20% headroom

	return valueobjects.ResourceRecommendation{
		ResourceType:     "memory_limit",
		CurrentValue:     p99Memory,
		RecommendedValue: recommended,
		HeadroomPercent:  20,
		Confidence:       valueobjects.ConfidenceMedium,
	}
}

// GenerateBatchRecommendations generates batch recommendations
func (r *rightSizer) GenerateBatchRecommendations(containers []*entities.Container, percentiles map[string]valueobjects.PercentileData) ([]*valueobjects.Recommendation, error) {
	// This would be implemented with concurrent processing
	// For now, return empty slice
	return make([]*valueobjects.Recommendation, 0), nil
}

// CalculateWasteGap calculates the waste gap between requested and utilized resources
func (r *rightSizer) CalculateWasteGap(container *entities.Container, utilizationPercentiles valueobjects.PercentileData) (cpuWaste, memoryWaste float64) {
	requestCPU := float64(container.Resources.CPURequest.MilliCPU) / 1000
	requestMemory := float64(container.Resources.MemoryRequest.Bytes)

	// Calculate CPU waste
	if requestCPU > 0 {
		p95CPU := utilizationPercentiles.P95CPU
		if p95CPU < requestCPU {
			cpuWaste = ((requestCPU - p95CPU) / requestCPU) * 100
		}
	}

	// Calculate memory waste
	if requestMemory > 0 {
		p95Memory := utilizationPercentiles.P95Memory
		if p95Memory < requestMemory {
			memoryWaste = ((requestMemory - p95Memory) / requestMemory) * 100
		}
	}

	return math.Max(0, cpuWaste), math.Max(0, memoryWaste)
}
