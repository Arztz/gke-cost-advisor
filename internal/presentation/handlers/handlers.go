package handlers

import (
	"encoding/json"
	"net/http"

	"gke-cost-advisor/internal/application/usecases"
	"gke-cost-advisor/internal/domain/services"
	"gke-cost-advisor/internal/domain/valueobjects"
	"gke-cost-advisor/internal/infrastructure/repositories"
)

// ClusterHandler handles cluster-related HTTP requests
type ClusterHandler struct {
	clusterRepo *repositories.ClusterRepository
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(clusterRepo *repositories.ClusterRepository) *ClusterHandler {
	return &ClusterHandler{clusterRepo: clusterRepo}
}

// ListClusters handles GET /api/v1/clusters
func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := h.clusterRepo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"clusters": clusters,
	})
}

// EfficiencyHandler handles efficiency analysis HTTP requests
type EfficiencyHandler struct {
	useCase *usecases.AnalyzeCostEfficiencyUseCase
}

// NewEfficiencyHandler creates a new efficiency handler
func NewEfficiencyHandler(useCase *usecases.AnalyzeCostEfficiencyUseCase) *EfficiencyHandler {
	return &EfficiencyHandler{useCase: useCase}
}

// AnalyzeEfficiency handles GET /api/v1/clusters/{id}/efficiency
func (h *EfficiencyHandler) AnalyzeEfficiency(w http.ResponseWriter, r *http.Request) {
	clusterID := r.PathValue("id")
	if clusterID == "" {
		http.Error(w, "cluster ID is required", http.StatusBadRequest)
		return
	}

	windowConfig := valueobjects.WindowConfig{
		Size:    24 * 3600 * 1000000000, // 24 hours in nanoseconds
		Step:    30 * 60 * 1000000000,
		Percent: 95,
	}

	req := usecases.AnalyzeCostEfficiencyRequest{
		ClusterID:    clusterID,
		WindowConfig: windowConfig,
	}

	result, err := h.useCase.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RecommendationHandler handles recommendation HTTP requests
type RecommendationHandler struct {
	rightSizer    services.RightSizer
	metricsRepo   *repositories.MetricsRepository
	pricingRepo   *repositories.PricingRepository
	containerRepo *repositories.ContainerRepository
}

// NewRecommendationHandler creates a new recommendation handler
func NewRecommendationHandler(rightSizer services.RightSizer, metricsRepo *repositories.MetricsRepository, pricingRepo *repositories.PricingRepository, containerRepo *repositories.ContainerRepository) *RecommendationHandler {
	return &RecommendationHandler{
		rightSizer:    rightSizer,
		metricsRepo:   metricsRepo,
		pricingRepo:   pricingRepo,
		containerRepo: containerRepo,
	}
}

// ListRecommendations handles GET /api/v1/recommendations
func (h *RecommendationHandler) ListRecommendations(w http.ResponseWriter, r *http.Request) {
	clusterID := r.URL.Query().Get("cluster_id")
	if clusterID == "" {
		http.Error(w, "cluster_id is required", http.StatusBadRequest)
		return
	}

	// Get containers for the cluster
	containers, err := h.containerRepo.GetAll(r.Context(), clusterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate sample recommendations
	recommendations := make([]*valueobjects.Recommendation, 0)
	for _, container := range containers {
		// Get percentiles
		percentiles, _ := h.metricsRepo.GetContainerMetrics(r.Context(), clusterID, container.ID, valueobjects.WindowConfig{Size: 24 * 3600 * 1000000000})

		// Calculate waste
		cpuWaste, memWaste := h.rightSizer.CalculateWasteGap(container, percentiles[container.ID])

		if cpuWaste > 10 || memWaste > 10 {
			rec := valueobjects.NewRecommendation(
				"rec-"+container.ID,
				clusterID,
				valueobjects.RecommendationTypeRightSizing,
				valueobjects.TargetTypeContainer,
				container.ID,
				container.Name,
				"current",
				"recommended",
				valueobjects.ConfidenceMedium,
				"Resource optimization opportunity detected",
			)
			recommendations = append(recommendations, rec)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
		"total":           len(recommendations),
	})
}

// MachineFamilyHandler handles machine family comparison HTTP requests
type MachineFamilyHandler struct {
	pricingRepo *repositories.PricingRepository
}

// NewMachineFamilyHandler creates a new machine family handler
func NewMachineFamilyHandler(pricingRepo *repositories.PricingRepository) *MachineFamilyHandler {
	return &MachineFamilyHandler{pricingRepo: pricingRepo}
}

// CompareMachineFamilies handles POST /api/v1/machine-families/compare
func (h *MachineFamilyHandler) CompareMachineFamilies(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Requirements struct {
			VCPU      int `json:"vcpu"`
			MemoryGB  int `json:"memory_gb"`
			StorageGB int `json:"storage_gb"`
		} `json:"requirements"`
		Region           string `json:"region"`
		IncludeSpot      bool   `json:"include_spot"`
		LatencySensitive bool   `json:"latency_sensitive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Region == "" {
		req.Region = "us-central1"
	}

	// Get all machine type pricing
	pricing, err := h.pricingRepo.GetAllMachineTypePricing(r.Context(), req.Region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build comparisons
	comparisons := make([]valueobjects.MachineFamily, 0)
	for machineType, info := range pricing {
		family := valueobjects.MachineFamily{
			MachineFamily:         info.MachineFamily,
			MachineType:           machineType,
			OnDemandHourly:        info.OnDemandHourly,
			SpotHourly:            info.SpotHourly,
			SpotSavingsPercent:    info.SpotSavingsPercent,
			PerformanceScore:      85,
			PricePerformanceRatio: 2.5,
			Recommended:           machineType == "e2-standard-4",
		}
		comparisons = append(comparisons, family)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comparisons": comparisons,
	})
}

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(mux *http.ServeMux, clusterHandler *ClusterHandler, efficiencyHandler *EfficiencyHandler, recommendationHandler *RecommendationHandler, machineFamilyHandler *MachineFamilyHandler) {
	// Cluster routes
	mux.HandleFunc("GET /api/v1/clusters", clusterHandler.ListClusters)
	mux.HandleFunc("GET /api/v1/clusters/{id}", clusterHandler.ListClusters)

	// Efficiency routes
	mux.HandleFunc("GET /api/v1/clusters/{id}/efficiency", efficiencyHandler.AnalyzeEfficiency)

	// Recommendation routes
	mux.HandleFunc("GET /api/v1/recommendations", recommendationHandler.ListRecommendations)

	// Machine family routes
	mux.HandleFunc("POST /api/v1/machine-families/compare", machineFamilyHandler.CompareMachineFamilies)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
