package usecases

import (
	"context"
	"fmt"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/valueobjects"
	"gke-cost-advisor/pkg/errors"
)

// GenerateRightSizingRecommendationsRequest represents a request to generate right-sizing recommendations
type GenerateRightSizingRecommendationsRequest struct {
	ClusterID      string
	Namespace      string
	WindowConfig   valueobjects.WindowConfig
	LatencyProfile valueobjects.LatencyProfile
}

// RightSizingRecommendationResult represents a right-sizing recommendation
type RightSizingRecommendationResult struct {
	ContainerID       string
	ContainerName     string
	Namespace         string
	CurrentCPURequest float64
	CurrentMemRequest float64
	RecommendedCPU    float64
	RecommendedMem    float64
	Confidence        valueobjects.ConfidenceLevel
	Savings           *valueobjects.Savings
}

// ContainerRepository provides container data access methods
type ContainerRepository interface {
	GetByNamespace(ctx context.Context, clusterID, namespace string) ([]*entities.Container, error)
}

// PricingRepository provides pricing data access methods
type PricingRepository interface {
	GetMachineTypePricing(ctx context.Context, machineType, region string) (*valueobjects.PricingInfo, error)
	GetAllMachineTypePricing(ctx context.Context, region string) (map[string]*valueobjects.PricingInfo, error)
}

// RightSizerService provides right-sizing operations
type RightSizerService interface {
	GenerateCPURequestRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation
	GenerateMemoryRequestRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation
	CalculateWasteGap(container *entities.Container, utilizationPercentiles valueobjects.PercentileData) (cpuWaste, memoryWaste float64)
}

// MetricsProvider provides metrics data access
type MetricsProvider interface {
	GetContainerMetrics(ctx context.Context, clusterID, containerID string, windowConfig valueobjects.WindowConfig) (map[string]valueobjects.PercentileData, error)
}

// GenerateRightSizingRecommendationsUseCase implements right-sizing recommendations
type GenerateRightSizingRecommendationsUseCase struct {
	rightSizer    RightSizerService
	metricsRepo   MetricsProvider
	containerRepo ContainerRepository
	pricingRepo   PricingRepository
}

// NewGenerateRightSizingRecommendationsUseCase creates a new use case
func NewGenerateRightSizingRecommendationsUseCase(
	rightSizer RightSizerService,
	metricsRepo MetricsProvider,
	containerRepo ContainerRepository,
	pricingRepo PricingRepository,
) *GenerateRightSizingRecommendationsUseCase {
	return &GenerateRightSizingRecommendationsUseCase{
		rightSizer:    rightSizer,
		metricsRepo:   metricsRepo,
		containerRepo: containerRepo,
		pricingRepo:   pricingRepo,
	}
}

// Execute generates right-sizing recommendations
func (uc *GenerateRightSizingRecommendationsUseCase) Execute(ctx context.Context, req GenerateRightSizingRecommendationsRequest) ([]RightSizingRecommendationResult, error) {
	if req.ClusterID == "" {
		return nil, errors.NewApplicationError("cluster ID is required", errors.ErrInvalidInput)
	}

	containers, err := uc.containerRepo.GetByNamespace(ctx, req.ClusterID, req.Namespace)
	if err != nil {
		return nil, err
	}

	results := make([]RightSizingRecommendationResult, 0, len(containers))
	for _, container := range containers {
		percentiles, err := uc.metricsRepo.GetContainerMetrics(ctx, req.ClusterID, container.ID, req.WindowConfig)
		if err != nil {
			continue
		}

		percentile, ok := percentiles[container.ID]
		if !ok {
			continue
		}

		cpuRec := uc.rightSizer.GenerateCPURequestRecommendation(percentile, req.LatencyProfile)
		memRec := uc.rightSizer.GenerateMemoryRequestRecommendation(percentile, req.LatencyProfile)

		cpuWaste, memWaste := uc.rightSizer.CalculateWasteGap(container, percentile)

		// Only include if there's meaningful waste
		if cpuWaste > 10 || memWaste > 10 {
			results = append(results, RightSizingRecommendationResult{
				ContainerID:       container.ID,
				ContainerName:     container.Name,
				Namespace:         container.Namespace,
				CurrentCPURequest: container.GetCPURequestCores(),
				CurrentMemRequest: float64(container.GetMemoryRequestBytes()) / (1024 * 1024 * 1024),
				RecommendedCPU:    cpuRec.RecommendedValue,
				RecommendedMem:    memRec.RecommendedValue / (1024 * 1024 * 1024),
				Confidence:        cpuRec.Confidence,
			})
		}
	}

	return results, nil
}

// CalculateFinancialImpactRequest represents a request to calculate financial impact
type CalculateFinancialImpactRequest struct {
	ClusterID       string
	Recommendations []RightSizingRecommendationResult
}

// CalculateFinancialImpactUseCase implements financial impact calculation
type CalculateFinancialImpactUseCase struct {
	pricingRepo PricingRepository
}

// NewCalculateFinancialImpactUseCase creates a new use case
func NewCalculateFinancialImpactUseCase(pricingRepo PricingRepository) *CalculateFinancialImpactUseCase {
	return &CalculateFinancialImpactUseCase{
		pricingRepo: pricingRepo,
	}
}

// Execute calculates financial impact of recommendations
func (uc *CalculateFinancialImpactUseCase) Execute(ctx context.Context, req CalculateFinancialImpactRequest) (*valueobjects.Savings, error) {
	if len(req.Recommendations) == 0 {
		return valueobjects.NewSavings(0), nil
	}

	var totalMonthlySavings float64

	for _, rec := range req.Recommendations {
		// Calculate CPU savings (difference in cores * hourly rate)
		cpuDiff := rec.CurrentCPURequest - rec.RecommendedCPU
		if cpuDiff > 0 {
			// Assume e2-standard-4 pricing as baseline
			pricing, err := uc.pricingRepo.GetMachineTypePricing(ctx, "e2-standard-4", "us-central1")
			if err == nil && pricing != nil {
				cpuSavings := cpuDiff * pricing.OnDemandHourly * 730 // hours per month
				totalMonthlySavings += cpuSavings
			}
		}

		// Calculate memory savings
		memDiff := (rec.CurrentMemRequest - rec.RecommendedMem) * 1024 * 1024 * 1024 // Convert GB to bytes
		if memDiff > 0 {
			pricing, err := uc.pricingRepo.GetMachineTypePricing(ctx, "e2-standard-4", "us-central1")
			if err == nil && pricing != nil {
				memSavings := float64(memDiff) / (1024 * 1024 * 1024) * pricing.PersistentDiskCost
				totalMonthlySavings += memSavings
			}
		}
	}

	savings := valueobjects.NewSavings(totalMonthlySavings)
	return savings, nil
}

// CompareMachineFamiliesRequest represents a request to compare machine families
type CompareMachineFamiliesRequest struct {
	WorkloadProfile  valueobjects.WorkloadProfile
	Region           string
	IncludeSpot      bool
	LatencySensitive bool
}

// CompareMachineFamiliesUseCase implements machine family comparison
type CompareMachineFamiliesUseCase struct {
	pricingRepo PricingRepository
}

// NewCompareMachineFamiliesUseCase creates a new use case
func NewCompareMachineFamiliesUseCase(pricingRepo PricingRepository) *CompareMachineFamiliesUseCase {
	return &CompareMachineFamiliesUseCase{
		pricingRepo: pricingRepo,
	}
}

// Execute compares machine families for a workload profile
func (uc *CompareMachineFamiliesUseCase) Execute(ctx context.Context, req CompareMachineFamiliesRequest) ([]valueobjects.MachineFamily, error) {
	if req.Region == "" {
		req.Region = "us-central1"
	}

	// Get all machine type pricing
	pricing, err := uc.pricingRepo.GetAllMachineTypePricing(ctx, req.Region)
	if err != nil {
		return nil, err
	}

	// Filter and score machine types
	comparisons := make([]valueobjects.MachineFamily, 0)

	for machineType, info := range pricing {
		// Skip if doesn't meet workload requirements
		if !meetsWorkloadRequirements(machineType, req.WorkloadProfile) {
			continue
		}

		// Calculate performance score based on workload profile
		perfScore := calculatePerformanceScore(machineType, req.WorkloadProfile, req.LatencySensitive)

		// Calculate price-performance ratio
		pricePerfRatio := calculatePricePerformanceRatio(info.OnDemandHourly, perfScore)

		family := valueobjects.MachineFamily{
			MachineFamily:         info.MachineFamily,
			MachineType:           machineType,
			OnDemandHourly:        info.OnDemandHourly,
			SpotHourly:            info.SpotHourly,
			SpotSavingsPercent:    info.SpotSavingsPercent,
			PerformanceScore:      perfScore,
			PricePerformanceRatio: pricePerfRatio,
			Recommended:           false,
		}

		comparisons = append(comparisons, family)
	}

	// Mark recommended (highest price-performance ratio for the requirements)
	if len(comparisons) > 0 {
		bestIdx := 0
		bestRatio := comparisons[0].PricePerformanceRatio
		for i, c := range comparisons {
			if c.PricePerformanceRatio > bestRatio {
				bestRatio = c.PricePerformanceRatio
				bestIdx = i
			}
		}
		comparisons[bestIdx].Recommended = true
	}

	return comparisons, nil
}

// meetsWorkloadRequirements checks if a machine type meets the workload profile
func meetsWorkloadRequirements(machineType string, profile valueobjects.WorkloadProfile) bool {
	var vcpu int
	fmt.Sscanf(machineType, "%*[a-z]-%*[a-z]-%d", &vcpu)

	memGB := vcpu * 4
	if containsString(machineType, "highmem") {
		memGB = vcpu * 8
	} else if containsString(machineType, "highcpu") {
		memGB = vcpu * 1
	}

	return vcpu >= profile.VCPU && memGB >= profile.MemoryGB
}

// calculatePerformanceScore calculates performance score for a machine type
func calculatePerformanceScore(machineType string, profile valueobjects.WorkloadProfile, latencySensitive bool) int {
	baseScore := 70

	switch {
	case containsString(machineType, "c3"):
		baseScore = 95 // Compute optimized
	case containsString(machineType, "n2"):
		baseScore = 85 // General purpose
	case containsString(machineType, "e2"):
		baseScore = 75 // Cost optimized
	}

	if latencySensitive && !containsString(machineType, "e2") {
		baseScore += 5
	}

	return baseScore
}

// calculatePricePerformanceRatio calculates price-performance ratio
func calculatePricePerformanceRatio(price float64, performance int) float64 {
	if price <= 0 {
		return 0
	}
	return float64(performance) / (price * 10)
}

// containsString is a simple string contains check
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
