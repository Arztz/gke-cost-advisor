package valueobjects

import (
	"time"
)

// ConfidenceLevel represents the confidence level of a calculation
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

// ScoreColor represents the color code for efficiency scores
type ScoreColor string

const (
	ScoreColorGreen  ScoreColor = "GREEN"
	ScoreColorYellow ScoreColor = "YELLOW"
	ScoreColorRed    ScoreColor = "RED"
)

// TimeRange represents a time range for analysis
type TimeRange struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// NewTimeRange creates a new TimeRange
func NewTimeRange(start, end time.Time) TimeRange {
	return TimeRange{
		StartTime: start,
		EndTime:   end,
		Duration:  end.Sub(start),
	}
}

// CostEfficiencyScore represents a calculated efficiency score
type CostEfficiencyScore struct {
	CompositeScore    float64         `json:"composite_score"`    // 0-100 scale
	CPUUtilization    float64         `json:"cpu_utilization"`    // 0-100 percentage
	MemoryUtilization float64         `json:"memory_utilization"` // 0-100 percentage
	StorageEfficiency float64         `json:"storage_efficiency"` // 0-100 percentage
	TimeRange         TimeRange       `json:"time_range"`
	Confidence        ConfidenceLevel `json:"confidence"`
	ColorCode         ScoreColor      `json:"color_code"`
}

// NewCostEfficiencyScore creates a new CostEfficiencyScore
func NewCostEfficiencyScore(cpu, memory, storage float64, timeRange TimeRange) CostEfficiencyScore {
	// Weighted formula: 30% CPU, 40% Memory, 30% Storage
	composite := (cpu * 0.30) + (memory * 0.40) + (storage * 0.30)

	// Determine color code
	colorCode := ScoreColorRed
	if composite >= 80 {
		colorCode = ScoreColorGreen
	} else if composite >= 50 {
		colorCode = ScoreColorYellow
	}

	// Determine confidence based on data quality
	confidence := ConfidenceMedium
	if cpu > 0 && memory > 0 && storage > 0 {
		confidence = ConfidenceHigh
	}

	return CostEfficiencyScore{
		CompositeScore:    composite,
		CPUUtilization:    cpu,
		MemoryUtilization: memory,
		StorageEfficiency: storage,
		TimeRange:         timeRange,
		Confidence:        confidence,
		ColorCode:         colorCode,
	}
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeRightSizing   RecommendationType = "right_sizing"
	RecommendationTypeSpotMigration RecommendationType = "spot_migration"
	RecommendationTypeNodePool      RecommendationType = "node_pool"
	RecommendationTypeLimitAdjust   RecommendationType = "limit_adjust"
)

// TargetType represents the type of target for a recommendation
type TargetType string

const (
	TargetTypeContainer TargetType = "container"
	TargetTypeWorkload  TargetType = "workload"
	TargetTypeNodePool  TargetType = "node_pool"
)

// RecommendationStatus represents the status of a recommendation
type RecommendationStatus string

const (
	RecommendationStatusPending  RecommendationStatus = "pending"
	RecommendationStatusAccepted RecommendationStatus = "accepted"
	RecommendationStatusRejected RecommendationStatus = "rejected"
	RecommendationStatusExpired  RecommendationStatus = "expired"
)

// Recommendation represents a single optimization recommendation
type Recommendation struct {
	ID                 string               `json:"id"`
	ClusterID          string               `json:"cluster_id"`
	RecommendationType RecommendationType   `json:"recommendation_type"`
	TargetType         TargetType           `json:"target_type"`
	TargetID           string               `json:"target_id"`
	TargetName         string               `json:"target_name"`
	CurrentValue       string               `json:"current_value"`
	RecommendedValue   string               `json:"recommended_value"`
	Confidence         ConfidenceLevel      `json:"confidence"`
	Savings            *Savings             `json:"savings,omitempty"`
	Justification      string               `json:"justification"`
	Status             RecommendationStatus `json:"status"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// NewRecommendation creates a new Recommendation
func NewRecommendation(
	id, clusterID string,
	recType RecommendationType,
	targetType TargetType,
	targetID, targetName, currentValue, recommendedValue string,
	confidence ConfidenceLevel,
	justification string,
) *Recommendation {
	now := time.Now()
	return &Recommendation{
		ID:                 id,
		ClusterID:          clusterID,
		RecommendationType: recType,
		TargetType:         targetType,
		TargetID:           targetID,
		TargetName:         targetName,
		CurrentValue:       currentValue,
		RecommendedValue:   recommendedValue,
		Confidence:         confidence,
		Justification:      justification,
		Status:             RecommendationStatusPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// SetSavings sets the savings for the recommendation
func (r *Recommendation) SetSavings(savings *Savings) {
	r.Savings = savings
}

// Accept marks the recommendation as accepted
func (r *Recommendation) Accept() {
	r.Status = RecommendationStatusAccepted
	r.UpdatedAt = time.Now()
}

// Reject marks the recommendation as rejected
func (r *Recommendation) Reject() {
	r.Status = RecommendationStatusRejected
	r.UpdatedAt = time.Now()
}

// Savings represents financial impact calculations
type Savings struct {
	MonthlySavings      float64 `json:"monthly_savings"`      // USD
	AnnualSavings       float64 `json:"annual_savings"`       // USD
	PercentageReduction float64 `json:"percentage_reduction"` // Percentage
	ROIMonths           int     `json:"roi_months"`           // Months to break-even
}

// NewSavings creates a new Savings value object
func NewSavings(monthlySavings float64) *Savings {
	return &Savings{
		MonthlySavings:      monthlySavings,
		AnnualSavings:       monthlySavings * 12,
		PercentageReduction: 0, // Will be calculated when current cost is known
		ROIMonths:           0,
	}
}

// CalculateWithCurrentCost calculates savings with current cost context
func (s *Savings) CalculateWithCurrentCost(currentMonthlyCost float64) {
	if currentMonthlyCost > 0 {
		s.PercentageReduction = (s.MonthlySavings / currentMonthlyCost) * 100
	}
}

// PricingType represents the type of pricing
type PricingType string

const (
	PricingTypeOnDemand  PricingType = "on_demand"
	PricingTypeSpot      PricingType = "spot"
	PricingTypeCommitted PricingType = "committed"
)

// PricingInfo represents Google Cloud pricing data
type PricingInfo struct {
	MachineType        string      `json:"machine_type"`
	MachineFamily      string      `json:"machine_family"`
	Region             string      `json:"region"`
	PricingType        PricingType `json:"pricing_type"`
	OnDemandHourly     float64     `json:"on_demand_hourly"`     // USD per hour
	SpotHourly         float64     `json:"spot_hourly"`          // USD per hour
	SpotSavingsPercent float64     `json:"spot_savings_percent"` // Percentage (60-91%)
	CommittedHourly    float64     `json:"committed_hourly"`     // USD per hour (if applicable)
	PersistentDiskCost float64     `json:"persistent_disk_cost"` // USD per GB per month
	Currency           string      `json:"currency"`             // Default: USD
	UpdatedAt          time.Time   `json:"updated_at"`
}

// NewPricingInfo creates a new PricingInfo
func NewPricingInfo(machineType, machineFamily, region string, onDemand, spot float64) *PricingInfo {
	spotSavings := 0.0
	if onDemand > 0 {
		spotSavings = ((onDemand - spot) / onDemand) * 100
	}

	return &PricingInfo{
		MachineType:        machineType,
		MachineFamily:      machineFamily,
		Region:             region,
		PricingType:        PricingTypeOnDemand,
		OnDemandHourly:     onDemand,
		SpotHourly:         spot,
		SpotSavingsPercent: spotSavings,
		Currency:           "USD",
		UpdatedAt:          time.Now(),
	}
}

// GetMonthlyCost returns monthly cost based on pricing type
func (p *PricingInfo) GetMonthlyCost(hoursPerMonth float64) float64 {
	switch p.PricingType {
	case PricingTypeSpot:
		return p.SpotHourly * hoursPerMonth
	case PricingTypeCommitted:
		return p.CommittedHourly * hoursPerMonth
	default:
		return p.OnDemandHourly * hoursPerMonth
	}
}

// WorkloadProfile represents workload requirements for machine family comparison
type WorkloadProfile struct {
	VCPU             int  `json:"vcpu"`              // Required vCPUs
	MemoryGB         int  `json:"memory_gb"`         // Required memory in GB
	StorageGB        int  `json:"storage_gb"`        // Required storage in GB
	LatencySensitive bool `json:"latency_sensitive"` // Is workload latency sensitive
}

// MachineFamily represents a machine family comparison result
type MachineFamily struct {
	MachineFamily         string  `json:"machine_family"`          // E2, N2, C3
	MachineType           string  `json:"machine_type"`            // e2-standard-4, etc.
	OnDemandHourly        float64 `json:"on_demand_hourly"`        // USD per hour
	SpotHourly            float64 `json:"spot_hourly"`             // USD per hour
	SpotSavingsPercent    float64 `json:"spot_savings_percent"`    // Percentage
	PerformanceScore      int     `json:"performance_score"`       // 0-100
	PricePerformanceRatio float64 `json:"price_performance_ratio"` // Performance per USD
	Recommended           bool    `json:"recommended"`             // Is recommended
}

// PercentileData represents percentile utilization data
type PercentileData struct {
	P50CPU    float64 `json:"p50_cpu"`    // 50th percentile CPU
	P90CPU    float64 `json:"p90_cpu"`    // 90th percentile CPU
	P95CPU    float64 `json:"p95_cpu"`    // 95th percentile CPU
	P99CPU    float64 `json:"p99_cpu"`    // 99th percentile CPU
	P50Memory float64 `json:"p50_memory"` // 50th percentile memory
	P90Memory float64 `json:"p90_memory"` // 90th percentile memory
	P95Memory float64 `json:"p95_memory"` // 95th percentile memory
	P99Memory float64 `json:"p99_memory"` // 99th percentile memory
}

// LatencyProfile represents latency sensitivity profile
type LatencyProfile string

const (
	LatencyProfileLow    LatencyProfile = "low"
	LatencyProfileMedium LatencyProfile = "medium"
	LatencyProfileHigh   LatencyProfile = "high"
)

// ResourceRecommendation represents a resource recommendation
type ResourceRecommendation struct {
	ResourceType     string          `json:"resource_type"`     // cpu_request, memory_limit, etc.
	CurrentValue     float64         `json:"current_value"`     // Current value
	RecommendedValue float64         `json:"recommended_value"` // Recommended value
	HeadroomPercent  float64         `json:"headroom_percent"`  // Headroom percentage
	Confidence       ConfidenceLevel `json:"confidence"`
}

// WindowConfig represents sliding window configuration
type WindowConfig struct {
	Size    time.Duration `json:"size"`    // Window size (1h, 6h, 24h, 7d)
	Step    time.Duration `json:"step"`    // Step interval
	Percent int           `json:"percent"` // Percentile to calculate (50, 90, 95, 99)
}

// DefaultWindowConfigs returns default window configurations
func DefaultWindowConfigs() map[string]WindowConfig {
	return map[string]WindowConfig{
		"1h":  {Size: time.Hour, Step: 5 * time.Minute, Percent: 95},
		"6h":  {Size: 6 * time.Hour, Step: 15 * time.Minute, Percent: 95},
		"24h": {Size: 24 * time.Hour, Step: 30 * time.Minute, Percent: 95},
		"7d":  {Size: 7 * 24 * time.Hour, Step: 4 * time.Hour, Percent: 95},
	}
}

// OOMRiskLevel represents the OOM risk level
type OOMRiskLevel string

const (
	OOMRiskLevelLow      OOMRiskLevel = "low"      // less than 60%
	OOMRiskLevelMedium   OOMRiskLevel = "medium"   // 60-80%
	OOMRiskLevelHigh     OOMRiskLevel = "high"     // greater than 80%
	OOMRiskLevelCritical OOMRiskLevel = "critical" // greater than 90%
)

// ThrottlingSeverity represents throttling severity
type ThrottlingSeverity string

const (
	ThrottlingSeverityNormal   ThrottlingSeverity = "normal"   // less than 10%
	ThrottlingSeverityWarning  ThrottlingSeverity = "warning"  // 10-25%
	ThrottlingSeverityCritical ThrottlingSeverity = "critical" // greater than 25%
)
