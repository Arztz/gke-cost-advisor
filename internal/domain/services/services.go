package services

import (
	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/valueobjects"
)

// CostCalculator defines the interface for cost calculation operations
type CostCalculator interface {
	CalculateCurrentCost(resources entities.ResourceRequirements, pricing *valueobjects.PricingInfo) float64
	CalculateProjectedCost(resources entities.ResourceRequirements, pricing *valueobjects.PricingInfo, strategy OptimizationStrategy) float64
	CalculateSpotSavings(onDemandCost float64, spotCost float64) *valueobjects.Savings
	CalculateAnnualProjection(monthlyCost float64) float64
	CalculateNodePoolCost(machineType string, nodeCount int, pricing *valueobjects.PricingInfo) float64
}

// OptimizationStrategy represents the optimization strategy to apply
type OptimizationStrategy string

const (
	StrategyOnDemand  OptimizationStrategy = "on_demand"
	StrategySpot      OptimizationStrategy = "spot"
	StrategyCommitted OptimizationStrategy = "committed"
	StrategyRightSize OptimizationStrategy = "right_size"
)

// EfficiencyScorer defines the interface for efficiency scoring operations
type EfficiencyScorer interface {
	CalculateNamespaceScore(metrics NamespaceMetrics) valueobjects.CostEfficiencyScore
	CalculateWorkloadScore(workload *entities.Workload, metrics NamespaceMetrics) valueobjects.CostEfficiencyScore
	CalculateClusterScore(cluster *entities.Cluster, metrics []NamespaceMetrics) valueobjects.CostEfficiencyScore
}

// NamespaceMetrics represents metrics for a namespace
type NamespaceMetrics struct {
	Namespace         string
	CPUUtilization    float64
	MemoryUtilization float64
	StorageEfficiency float64
	ContainerCount    int
	TimeRange         valueobjects.TimeRange
}

// PricingAnalyzer defines the interface for pricing analysis operations
type PricingAnalyzer interface {
	CompareMachineFamilies(workload valueobjects.WorkloadProfile, region string) ([]valueobjects.MachineFamily, error)
	CalculatePricePerformanceRatio(machineType string, workload valueobjects.WorkloadProfile, pricing *valueobjects.PricingInfo) float64
	RecommendMachineMigration(currentMachineType string, workload valueobjects.WorkloadProfile, region string) (*valueobjects.Recommendation, error)
	GetSpotSavingsPotential(onDemandCost float64, machineType string, region string) (*valueobjects.Savings, error)
}

// RightSizer defines the interface for right-sizing operations
type RightSizer interface {
	GenerateCPURequestRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation
	GenerateMemoryRequestRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation
	GenerateCPULimitRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation
	GenerateMemoryLimitRecommendation(utilizationPercentiles valueobjects.PercentileData, latencyProfile valueobjects.LatencyProfile) valueobjects.ResourceRecommendation
	GenerateBatchRecommendations(containers []*entities.Container, percentiles map[string]valueobjects.PercentileData) ([]*valueobjects.Recommendation, error)
	CalculateWasteGap(container *entities.Container, utilizationPercentiles valueobjects.PercentileData) (cpuWaste, memoryWaste float64)
}

// OOMAnalyzer defines the interface for OOM analysis operations
type OOMAnalyzer interface {
	AnalyzeOOMRisk(container *entities.Container, memoryWorkingSetBytes int64, timeRange valueobjects.TimeRange) OOMRiskAssessment
	PredictDaysToOOM(memoryWorkingSetTrend []DataPoint) int
	GetMemoryLeakPattern(memoryTrend []DataPoint) MemoryLeakPattern
}

// OOMRiskAssessment represents an OOM risk assessment
type OOMRiskAssessment struct {
	ContainerID         string
	CurrentUsagePercent float64
	RiskLevel           valueobjects.OOMRiskLevel
	Trend               TrendDirection
	DaysToOOM           int
	Recommendations     []string
}

// DataPoint represents a time-series data point
type DataPoint struct {
	Timestamp int64
	Value     float64
}

// TrendDirection represents the trend direction
type TrendDirection string

const (
	TrendIncreasing TrendDirection = "increasing"
	TrendStable     TrendDirection = "stable"
	TrendDecreasing TrendDirection = "decreasing"
)

// MemoryLeakPattern represents a memory leak pattern
type MemoryLeakPattern struct {
	IsDetected bool
	Severity   float64 // 0-100
	TrendRate  float64 // bytes per hour
	Confidence valueobjects.ConfidenceLevel
}

// ThrottlingAnalyzer defines the interface for throttling analysis
type ThrottlingAnalyzer interface {
	AnalyzeThrottling(containerID string, throttledSeconds float64, totalSeconds float64) ThrottlingAnalysis
	DetectPerformanceBottleneck(container *entities.Container, throttlingAnalysis ThrottlingAnalysis) *valueobjects.Recommendation
}

// ThrottlingAnalysis represents a throttling analysis result
type ThrottlingAnalysis struct {
	ContainerID       string
	ThrottledSeconds  float64
	TotalSeconds      float64
	ThrottlingPercent float64
	Severity          valueobjects.ThrottlingSeverity
	IsAtCPULimit      bool
}
