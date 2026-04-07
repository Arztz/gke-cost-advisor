package entities

import (
	"time"

	"gke-cost-advisor/pkg/errors"
)

// ResourceRequirements represents Kubernetes resource requests and limits
type ResourceRequirements struct {
	CPURequest    *ResourceQuantity `json:"cpu_request,omitempty"`
	CPULimit      *ResourceQuantity `json:"cpu_limit,omitempty"`
	MemoryRequest *ResourceQuantity `json:"memory_request,omitempty"`
	MemoryLimit   *ResourceQuantity `json:"memory_limit,omitempty"`
}

// ResourceQuantity represents a resource quantity (CPU in cores, Memory in bytes)
type ResourceQuantity struct {
	MilliCPU int64 // Millicores (1000m = 1 core)
	Bytes    int64 // Bytes
}

// NewResourceQuantity creates a new ResourceQuantity from millicores and bytes
func NewResourceQuantity(milliCPU int64, bytes int64) *ResourceQuantity {
	return &ResourceQuantity{
		MilliCPU: milliCPU,
		Bytes:    bytes,
	}
}

// Container represents a single container within a Kubernetes pod
type Container struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Namespace string               `json:"namespace"`
	PodName   string               `json:"pod_name"`
	Image     string               `json:"image"`
	Resources ResourceRequirements `json:"resources"`
	Labels    map[string]string    `json:"labels"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// NewContainer creates a new Container entity
func NewContainer(id, name, namespace, podName, image string, resources ResourceRequirements) *Container {
	now := time.Now()
	return &Container{
		ID:        id,
		Name:      name,
		Namespace: namespace,
		PodName:   podName,
		Image:     image,
		Resources: resources,
		Labels:    make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetCPURequestCores returns CPU request in cores
func (c *Container) GetCPURequestCores() float64 {
	if c.Resources.CPURequest == nil {
		return 0
	}
	return float64(c.Resources.CPURequest.MilliCPU) / 1000
}

// GetMemoryRequestBytes returns memory request in bytes
func (c *Container) GetMemoryRequestBytes() int64 {
	if c.Resources.MemoryRequest == nil {
		return 0
	}
	return c.Resources.MemoryRequest.Bytes
}

// Validate validates the container entity
func (c *Container) Validate() error {
	if c.ID == "" {
		return errors.NewDomainError("container ID is required", "CONTAINER_ID_REQUIRED")
	}
	if c.Name == "" {
		return errors.NewDomainError("container name is required", "CONTAINER_NAME_REQUIRED")
	}
	if c.Namespace == "" {
		return errors.NewDomainError("container namespace is required", "CONTAINER_NAMESPACE_REQUIRED")
	}
	return nil
}

// WorkloadType represents the type of Kubernetes workload
type WorkloadType string

const (
	WorkloadTypeDeployment  WorkloadType = "Deployment"
	WorkloadTypeStatefulSet WorkloadType = "StatefulSet"
	WorkloadTypeDaemonSet   WorkloadType = "DaemonSet"
	WorkloadTypeJob         WorkloadType = "Job"
	WorkloadTypeCronJob     WorkloadType = "CronJob"
)

// Workload represents a Kubernetes workload (Deployment, StatefulSet, etc.)
type Workload struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	WorkloadType WorkloadType      `json:"workload_type"`
	ReplicaCount int               `json:"replica_count"`
	Containers   []*Container      `json:"containers"`
	Labels       map[string]string `json:"labels"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewWorkload creates a new Workload entity
func NewWorkload(id, name, namespace string, workloadType WorkloadType, replicaCount int) *Workload {
	now := time.Now()
	return &Workload{
		ID:           id,
		Name:         name,
		Namespace:    namespace,
		WorkloadType: workloadType,
		ReplicaCount: replicaCount,
		Containers:   make([]*Container, 0),
		Labels:       make(map[string]string),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AddContainer adds a container to the workload
func (w *Workload) AddContainer(container *Container) {
	w.Containers = append(w.Containers, container)
}

// GetTotalCPURequest returns total CPU request across all containers
func (w *Workload) GetTotalCPURequest() float64 {
	var total float64
	for _, c := range w.Containers {
		total += c.GetCPURequestCores()
	}
	return total * float64(w.ReplicaCount)
}

// GetTotalMemoryRequest returns total memory request across all containers
func (w *Workload) GetTotalMemoryRequest() int64 {
	var total int64
	for _, c := range w.Containers {
		total += c.GetMemoryRequestBytes()
	}
	return total * int64(w.ReplicaCount)
}

// Validate validates the workload entity
func (w *Workload) Validate() error {
	if w.ID == "" {
		return errors.NewDomainError("workload ID is required", "WORKLOAD_ID_REQUIRED")
	}
	if w.Name == "" {
		return errors.NewDomainError("workload name is required", "WORKLOAD_NAME_REQUIRED")
	}
	if w.Namespace == "" {
		return errors.NewDomainError("workload namespace is required", "WORKLOAD_NAMESPACE_REQUIRED")
	}
	return nil
}

// NodePool represents a GKE node pool
type NodePool struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ClusterID   string    `json:"cluster_id"`
	MachineType string    `json:"machine_type"`
	NodeCount   int       `json:"node_count"`
	MinNodes    int       `json:"min_nodes"`
	MaxNodes    int       `json:"max_nodes"`
	AutoScaling bool      `json:"auto_scaling"`
	Zone        string    `json:"zone"`
	Preemptible bool      `json:"preemptible"`
	DiskSizeGB  int       `json:"disk_size_gb"`
	DiskType    string    `json:"disk_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewNodePool creates a new NodePool entity
func NewNodePool(id, name, clusterID, machineType, zone string, nodeCount int, autoScaling bool) *NodePool {
	now := time.Now()
	return &NodePool{
		ID:          id,
		Name:        name,
		ClusterID:   clusterID,
		MachineType: machineType,
		NodeCount:   nodeCount,
		MinNodes:    nodeCount,
		MaxNodes:    nodeCount,
		AutoScaling: autoScaling,
		Zone:        zone,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate validates the node pool entity
func (np *NodePool) Validate() error {
	if np.ID == "" {
		return errors.NewDomainError("node pool ID is required", "NODEPOOL_ID_REQUIRED")
	}
	if np.Name == "" {
		return errors.NewDomainError("node pool name is required", "NODEPOOL_NAME_REQUIRED")
	}
	if np.ClusterID == "" {
		return errors.NewDomainError("node pool cluster ID is required", "NODEPOOL_CLUSTER_ID_REQUIRED")
	}
	return nil
}

// Cluster represents a GKE cluster
type Cluster struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Region             string      `json:"region"`
	Location           string      `json:"location"`
	ProjectID          string      `json:"project_id"`
	MasterVersion      string      `json:"master_version"`
	NodePools          []*NodePool `json:"node_pools"`
	IsActive           bool        `json:"is_active"`
	KubeconfigPath     string      `json:"kubeconfig_path,omitempty"`
	PrometheusEndpoint string      `json:"prometheus_endpoint"`
	BillingAccountID   string      `json:"billing_account_id,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// NewCluster creates a new Cluster entity
func NewCluster(id, name, region, location, projectID, prometheusEndpoint string) *Cluster {
	now := time.Now()
	return &Cluster{
		ID:                 id,
		Name:               name,
		Region:             region,
		Location:           location,
		ProjectID:          projectID,
		NodePools:          make([]*NodePool, 0),
		IsActive:           true,
		PrometheusEndpoint: prometheusEndpoint,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// AddNodePool adds a node pool to the cluster
func (c *Cluster) AddNodePool(nodePool *NodePool) {
	c.NodePools = append(c.NodePools, nodePool)
}

// Validate validates the cluster entity
func (c *Cluster) Validate() error {
	if c.ID == "" {
		return errors.NewDomainError("cluster ID is required", "CLUSTER_ID_REQUIRED")
	}
	if c.Name == "" {
		return errors.NewDomainError("cluster name is required", "CLUSTER_NAME_REQUIRED")
	}
	if c.PrometheusEndpoint == "" {
		return errors.NewDomainError("prometheus endpoint is required", "CLUSTER_PROMETHEUS_REQUIRED")
	}
	return nil
}
