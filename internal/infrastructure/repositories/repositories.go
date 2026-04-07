package repositories

import (
	"context"
	"sync"
	"time"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/services"
	"gke-cost-advisor/internal/domain/valueobjects"
	"gke-cost-advisor/internal/infrastructure/clients"
	"gke-cost-advisor/pkg/config"
)

// ClusterRepository handles cluster data access
type ClusterRepository struct {
	kubernetesClient interface {
		ListNamespaces(ctx context.Context) ([]string, error)
		GetNodePools(ctx context.Context, cluster string) ([]*entities.NodePool, error)
	}
	config   *config.Config
	mu       sync.RWMutex
	clusters map[string]*entities.Cluster
}

// NewClusterRepository creates a new cluster repository
func NewClusterRepository(k8sClient interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	GetNodePools(ctx context.Context, cluster string) ([]*entities.NodePool, error)
}, cfg *config.Config) *ClusterRepository {
	return &ClusterRepository{
		kubernetesClient: k8sClient,
		config:           cfg,
		clusters:         make(map[string]*entities.Cluster),
	}
}

// GetByID retrieves a cluster by ID
func (r *ClusterRepository) GetByID(ctx context.Context, id string) (*entities.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cluster, ok := r.clusters[id]; ok {
		return cluster, nil
	}

	// Return a default cluster if not found (for demo purposes)
	cluster := entities.NewCluster(
		id,
		"demo-cluster",
		r.config.GCP.Region,
		r.config.GCP.Region+"-a",
		r.config.GCP.ProjectID,
		r.config.Prometheus.Endpoint,
	)
	cluster.BillingAccountID = r.config.GCP.BillingAccountID
	return cluster, nil
}

// List retrieves all clusters
func (r *ClusterRepository) List(ctx context.Context) ([]*entities.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusters := make([]*entities.Cluster, 0, len(r.clusters))
	for _, c := range r.clusters {
		clusters = append(clusters, c)
	}

	// Return demo cluster if none registered
	if len(clusters) == 0 {
		cluster := entities.NewCluster(
			"demo-cluster",
			"demo-cluster",
			r.config.GCP.Region,
			r.config.GCP.Region+"-a",
			r.config.GCP.ProjectID,
			r.config.Prometheus.Endpoint,
		)
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// ListNamespaces retrieves namespaces for a cluster
func (r *ClusterRepository) ListNamespaces(ctx context.Context, clusterID string) ([]string, error) {
	if r.kubernetesClient != nil {
		return r.kubernetesClient.ListNamespaces(ctx)
	}
	// Default namespaces
	return []string{"default", "kube-system", "production", "staging"}, nil
}

// Save saves a cluster
func (r *ClusterRepository) Save(ctx context.Context, cluster *entities.Cluster) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clusters[cluster.ID] = cluster
	return nil
}

// ContainerRepository handles container data access
type ContainerRepository struct {
	kubernetesClient interface{}
	mu               sync.RWMutex
	containers       map[string]map[string]*entities.Container
}

// NewContainerRepository creates a new container repository
func NewContainerRepository(k8sClient interface{}) *ContainerRepository {
	return &ContainerRepository{
		kubernetesClient: k8sClient,
		containers:       make(map[string]map[string]*entities.Container),
	}
}

// GetByIDs retrieves containers by IDs
func (r *ContainerRepository) GetByIDs(ctx context.Context, clusterID string, ids []string) ([]*entities.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusterContainers, ok := r.containers[clusterID]
	if !ok {
		return nil, nil
	}

	result := make([]*entities.Container, 0, len(ids))
	for _, id := range ids {
		if c, ok := clusterContainers[id]; ok {
			result = append(result, c)
		}
	}
	return result, nil
}

// GetByWorkload retrieves containers by workload
func (r *ContainerRepository) GetByWorkload(ctx context.Context, clusterID, workloadID string) ([]*entities.Container, error) {
	// Stub implementation
	return make([]*entities.Container, 0), nil
}

// GetByNamespace retrieves containers by namespace
func (r *ContainerRepository) GetByNamespace(ctx context.Context, clusterID, namespace string) ([]*entities.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusterContainers, ok := r.containers[clusterID]
	if !ok {
		// Return sample containers for demo
		return r.getSampleContainers(clusterID, namespace), nil
	}

	result := make([]*entities.Container, 0)
	for _, c := range clusterContainers {
		if c.Namespace == namespace {
			result = append(result, c)
		}
	}
	return result, nil
}

// GetAll retrieves all containers for a cluster
func (r *ContainerRepository) GetAll(ctx context.Context, clusterID string) ([]*entities.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusterContainers, ok := r.containers[clusterID]
	if !ok {
		return r.getSampleContainers(clusterID, ""), nil
	}

	result := make([]*entities.Container, 0, len(clusterContainers))
	for _, c := range clusterContainers {
		result = append(result, c)
	}
	return result, nil
}

// getSampleContainers returns sample containers for demo
func (r *ContainerRepository) getSampleContainers(clusterID, namespace string) []*entities.Container {
	containers := []*entities.Container{
		{
			ID:        "container-1",
			Name:      "api-server",
			Namespace: "production",
			PodName:   "api-server-0",
			Image:     "nginx:latest",
			Resources: entities.ResourceRequirements{
				CPURequest:    &entities.ResourceQuantity{MilliCPU: 1000},
				CPULimit:      &entities.ResourceQuantity{MilliCPU: 2000},
				MemoryRequest: &entities.ResourceQuantity{Bytes: 1024 * 1024 * 1024},
				MemoryLimit:   &entities.ResourceQuantity{Bytes: 2 * 1024 * 1024 * 1024},
			},
		},
		{
			ID:        "container-2",
			Name:      "worker",
			Namespace: "production",
			PodName:   "worker-0",
			Image:     "worker:latest",
			Resources: entities.ResourceRequirements{
				CPURequest:    &entities.ResourceQuantity{MilliCPU: 500},
				MemoryRequest: &entities.ResourceQuantity{Bytes: 512 * 1024 * 1024},
			},
		},
	}

	if namespace != "" {
		filtered := make([]*entities.Container, 0)
		for _, c := range containers {
			if c.Namespace == namespace {
				filtered = append(filtered, c)
			}
		}
		return filtered
	}

	return containers
}

// MetricsRepository handles metrics data access
type MetricsRepository struct {
	prometheusClient *clients.PrometheusClient
	config           *config.Config
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(promClient *clients.PrometheusClient, cfg *config.Config) *MetricsRepository {
	return &MetricsRepository{
		prometheusClient: promClient,
		config:           cfg,
	}
}

// GetNamespaceMetrics retrieves metrics for a namespace
func (r *MetricsRepository) GetNamespaceMetrics(ctx context.Context, clusterID, namespace string, timeRange valueobjects.TimeRange) (services.NamespaceMetrics, error) {
	// Stub implementation with sample data
	// In production, would query Prometheus

	now := time.Now()
	metrics := services.NamespaceMetrics{
		Namespace:         namespace,
		CPUUtilization:    65.5 + float64(len(namespace)*10),
		MemoryUtilization: 72.3 + float64(len(namespace)*5),
		StorageEfficiency: 68.0,
		ContainerCount:    15,
		TimeRange:         valueobjects.NewTimeRange(now.Add(-24*time.Hour), now),
	}

	return metrics, nil
}

// GetContainerMetrics retrieves container metrics
func (r *MetricsRepository) GetContainerMetrics(ctx context.Context, clusterID, containerID string, windowConfig valueobjects.WindowConfig) (map[string]valueobjects.PercentileData, error) {
	// Stub implementation with sample data
	percentiles := map[string]valueobjects.PercentileData{
		containerID: {
			P50CPU:    0.5,
			P90CPU:    0.8,
			P95CPU:    1.2,
			P99CPU:    1.5,
			P50Memory: 512 * 1024 * 1024,
			P90Memory: 768 * 1024 * 1024,
			P95Memory: 900 * 1024 * 1024,
			P99Memory: 1024 * 1024 * 1024,
		},
	}

	return percentiles, nil
}

// GetContainerThrottling retrieves container throttling metrics
func (r *MetricsRepository) GetContainerThrottling(ctx context.Context, clusterID, containerID string, timeRange valueobjects.TimeRange) (services.ThrottlingAnalysis, error) {
	// Stub implementation
	return services.ThrottlingAnalysis{
		ContainerID:       containerID,
		ThrottledSeconds:  120,
		TotalSeconds:      3600,
		ThrottlingPercent: 3.3,
		Severity:          valueobjects.ThrottlingSeverityNormal,
		IsAtCPULimit:      false,
	}, nil
}

// GetContainerMemoryWorkingSet retrieves container memory working set
func (r *MetricsRepository) GetContainerMemoryWorkingSet(ctx context.Context, clusterID, containerID string, timeRange valueobjects.TimeRange) ([]services.DataPoint, error) {
	// Stub implementation
	now := time.Now().Unix()
	points := make([]services.DataPoint, 0, 10)
	for i := 0; i < 10; i++ {
		points = append(points, services.DataPoint{
			Timestamp: now - int64(i*3600),
			Value:     512 * 1024 * 1024,
		})
	}
	return points, nil
}

// PricingRepository handles pricing data access
type PricingRepository struct {
	billingClient *clients.BillingClient
	config        *config.Config
	mu            sync.RWMutex
	cache         map[string]*valueobjects.PricingInfo
}

// NewPricingRepository creates a new pricing repository
func NewPricingRepository(billingClient *clients.BillingClient, cfg *config.Config) *PricingRepository {
	return &PricingRepository{
		billingClient: billingClient,
		config:        cfg,
		cache:         make(map[string]*valueobjects.PricingInfo),
	}
}

// GetRegionalPricing retrieves regional pricing
func (r *PricingRepository) GetRegionalPricing(ctx context.Context, region string) (*valueobjects.PricingInfo, error) {
	r.mu.RLock()
	if pricing, ok := r.cache[region]; ok {
		r.mu.RUnlock()
		return pricing, nil
	}
	r.mu.RUnlock()

	// Get from client or use stub
	if r.billingClient != nil {
		pricing, err := r.billingClient.GetMachineTypePricing(ctx, "e2-standard-4", region)
		if err == nil {
			r.mu.Lock()
			r.cache[region] = pricing
			r.mu.Unlock()
			return pricing, nil
		}
	}

	// Return stub pricing
	pricing := valueobjects.NewPricingInfo("e2-standard-4", "E2", region, 0.192, 0.057)
	r.mu.Lock()
	r.cache[region] = pricing
	r.mu.Unlock()

	return pricing, nil
}

// GetMachineTypePricing retrieves pricing for a specific machine type
func (r *PricingRepository) GetMachineTypePricing(ctx context.Context, machineType, region string) (*valueobjects.PricingInfo, error) {
	cacheKey := machineType + "-" + region

	r.mu.RLock()
	if pricing, ok := r.cache[cacheKey]; ok {
		r.mu.RUnlock()
		return pricing, nil
	}
	r.mu.RUnlock()

	if r.billingClient != nil {
		pricing, err := r.billingClient.GetMachineTypePricing(ctx, machineType, region)
		if err == nil {
			r.mu.Lock()
			r.cache[cacheKey] = pricing
			r.mu.Unlock()
			return pricing, nil
		}
	}

	// Return stub pricing
	pricing := valueobjects.NewPricingInfo(machineType, "E2", region, 0.192, 0.057)
	r.mu.Lock()
	r.cache[cacheKey] = pricing
	r.mu.Unlock()

	return pricing, nil
}

// GetAllMachineTypePricing retrieves all machine type pricing
func (r *PricingRepository) GetAllMachineTypePricing(ctx context.Context, region string) (map[string]*valueobjects.PricingInfo, error) {
	if r.billingClient != nil {
		return r.billingClient.GetAllMachineTypePricing(ctx, region)
	}

	// Return stub data
	pricing := make(map[string]*valueobjects.PricingInfo)
	machineTypes := []string{"e2-standard-4", "n2-standard-4", "c3-standard-4", "e2-highcpu-4", "c3-highmem-4"}
	for _, mt := range machineTypes {
		pricing[mt] = valueobjects.NewPricingInfo(mt, "E2", region, 0.192, 0.057)
	}
	return pricing, nil
}

// GetSpotPricing retrieves Spot pricing
func (r *PricingRepository) GetSpotPricing(ctx context.Context, machineType, region string) (*valueobjects.PricingInfo, error) {
	pricing, err := r.GetMachineTypePricing(ctx, machineType, region)
	if err != nil {
		return nil, err
	}
	pricing.PricingType = valueobjects.PricingTypeSpot
	return pricing, nil
}
