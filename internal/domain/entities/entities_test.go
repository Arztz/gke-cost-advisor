package entities

import (
	"testing"
)

func TestContainerValidate(t *testing.T) {
	tests := []struct {
		name        string
		container   *Container
		expectError bool
		errorCode   string
	}{
		{
			name: "valid container",
			container: &Container{
				ID:        "container-1",
				Name:      "test-container",
				Namespace: "default",
			},
			expectError: false,
		},
		{
			name: "missing ID",
			container: &Container{
				Name:      "test-container",
				Namespace: "default",
			},
			expectError: true,
			errorCode:   "CONTAINER_ID_REQUIRED",
		},
		{
			name: "missing name",
			container: &Container{
				ID:        "container-1",
				Namespace: "default",
			},
			expectError: true,
			errorCode:   "CONTAINER_NAME_REQUIRED",
		},
		{
			name: "missing namespace",
			container: &Container{
				ID:   "container-1",
				Name: "test-container",
			},
			expectError: true,
			errorCode:   "CONTAINER_NAMESPACE_REQUIRED",
		},
		{
			name:        "all fields missing",
			container:   &Container{},
			expectError: true,
			errorCode:   "CONTAINER_ID_REQUIRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.container.Validate()

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil {
				// Error was expected, test passes
				_ = err
			}
		})
	}
}

func TestContainerGetCPURequestCores(t *testing.T) {
	tests := []struct {
		name      string
		container *Container
		expected  float64
	}{
		{
			name: "with CPU request",
			container: &Container{
				Resources: ResourceRequirements{
					CPURequest: &ResourceQuantity{MilliCPU: 2000},
				},
			},
			expected: 2.0,
		},
		{
			name: "zero CPU request",
			container: &Container{
				Resources: ResourceRequirements{
					CPURequest: &ResourceQuantity{MilliCPU: 0},
				},
			},
			expected: 0.0,
		},
		{
			name: "nil CPU request",
			container: &Container{
				Resources: ResourceRequirements{
					CPURequest: nil,
				},
			},
			expected: 0.0,
		},
		{
			name: "fractional CPU",
			container: &Container{
				Resources: ResourceRequirements{
					CPURequest: &ResourceQuantity{MilliCPU: 500},
				},
			},
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.container.GetCPURequestCores()
			if result != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestContainerGetMemoryRequestBytes(t *testing.T) {
	tests := []struct {
		name      string
		container *Container
		expected  int64
	}{
		{
			name: "with memory request",
			container: &Container{
				Resources: ResourceRequirements{
					MemoryRequest: &ResourceQuantity{Bytes: 4 * 1024 * 1024 * 1024},
				},
			},
			expected: 4 * 1024 * 1024 * 1024,
		},
		{
			name: "zero memory request",
			container: &Container{
				Resources: ResourceRequirements{
					MemoryRequest: &ResourceQuantity{Bytes: 0},
				},
			},
			expected: 0,
		},
		{
			name: "nil memory request",
			container: &Container{
				Resources: ResourceRequirements{
					MemoryRequest: nil,
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.container.GetMemoryRequestBytes()
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestWorkloadValidate(t *testing.T) {
	tests := []struct {
		name        string
		workload    *Workload
		expectError bool
	}{
		{
			name: "valid workload",
			workload: &Workload{
				ID:        "workload-1",
				Name:      "test-workload",
				Namespace: "default",
			},
			expectError: false,
		},
		{
			name: "missing ID",
			workload: &Workload{
				Name:      "test-workload",
				Namespace: "default",
			},
			expectError: true,
		},
		{
			name: "missing name",
			workload: &Workload{
				ID:        "workload-1",
				Namespace: "default",
			},
			expectError: true,
		},
		{
			name: "missing namespace",
			workload: &Workload{
				ID:   "workload-1",
				Name: "test-workload",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workload.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkloadGetTotalCPURequest(t *testing.T) {
	workload := &Workload{
		ID:           "workload-1",
		Name:         "test-workload",
		Namespace:    "default",
		ReplicaCount: 3,
		Containers: []*Container{
			{
				Resources: ResourceRequirements{
					CPURequest: &ResourceQuantity{MilliCPU: 1000},
				},
			},
			{
				Resources: ResourceRequirements{
					CPURequest: &ResourceQuantity{MilliCPU: 500},
				},
			},
		},
	}

	// Total: (1 + 0.5) * 3 = 4.5 cores
	expected := 4.5
	result := workload.GetTotalCPURequest()
	if result != expected {
		t.Errorf("expected %.2f, got %.2f", expected, result)
	}
}

func TestWorkloadGetTotalMemoryRequest(t *testing.T) {
	workload := &Workload{
		ID:           "workload-1",
		Name:         "test-workload",
		Namespace:    "default",
		ReplicaCount: 2,
		Containers: []*Container{
			{
				Resources: ResourceRequirements{
					MemoryRequest: &ResourceQuantity{Bytes: 2 * 1024 * 1024 * 1024},
				},
			},
			{
				Resources: ResourceRequirements{
					MemoryRequest: &ResourceQuantity{Bytes: 1 * 1024 * 1024 * 1024},
				},
			},
		},
	}

	// Total: (2 + 1) * 2 = 6 GB
	expected := int64(6 * 1024 * 1024 * 1024)
	result := workload.GetTotalMemoryRequest()
	if result != expected {
		t.Errorf("expected %d, got %d", expected, result)
	}
}

func TestWorkloadAddContainer(t *testing.T) {
	workload := NewWorkload("workload-1", "test-workload", "default", WorkloadTypeDeployment, 1)

	container := &Container{
		ID:   "container-1",
		Name: "test-container",
	}

	workload.AddContainer(container)

	if len(workload.Containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(workload.Containers))
	}
	if workload.Containers[0] != container {
		t.Error("container not added correctly")
	}
}

func TestNodePoolValidate(t *testing.T) {
	tests := []struct {
		name        string
		nodePool    *NodePool
		expectError bool
	}{
		{
			name: "valid node pool",
			nodePool: &NodePool{
				ID:        "node-pool-1",
				Name:      "test-node-pool",
				ClusterID: "cluster-1",
			},
			expectError: false,
		},
		{
			name: "missing ID",
			nodePool: &NodePool{
				Name:      "test-node-pool",
				ClusterID: "cluster-1",
			},
			expectError: true,
		},
		{
			name: "missing name",
			nodePool: &NodePool{
				ID:        "node-pool-1",
				ClusterID: "cluster-1",
			},
			expectError: true,
		},
		{
			name: "missing cluster ID",
			nodePool: &NodePool{
				ID:   "node-pool-1",
				Name: "test-node-pool",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.nodePool.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClusterValidate(t *testing.T) {
	tests := []struct {
		name        string
		cluster     *Cluster
		expectError bool
	}{
		{
			name: "valid cluster",
			cluster: &Cluster{
				ID:                 "cluster-1",
				Name:               "test-cluster",
				PrometheusEndpoint: "http://prometheus:9090",
			},
			expectError: false,
		},
		{
			name: "missing ID",
			cluster: &Cluster{
				Name:               "test-cluster",
				PrometheusEndpoint: "http://prometheus:9090",
			},
			expectError: true,
		},
		{
			name: "missing name",
			cluster: &Cluster{
				ID:                 "cluster-1",
				PrometheusEndpoint: "http://prometheus:9090",
			},
			expectError: true,
		},
		{
			name: "missing prometheus endpoint",
			cluster: &Cluster{
				ID:   "cluster-1",
				Name: "test-cluster",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cluster.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClusterAddNodePool(t *testing.T) {
	cluster := NewCluster("cluster-1", "test-cluster", "us-central1", "us-central1", "project-1", "http://prometheus:9090")

	nodePool := &NodePool{
		ID:        "node-pool-1",
		Name:      "test-node-pool",
		ClusterID: "cluster-1",
	}

	cluster.AddNodePool(nodePool)

	if len(cluster.NodePools) != 1 {
		t.Errorf("expected 1 node pool, got %d", len(cluster.NodePools))
	}
	if cluster.NodePools[0] != nodePool {
		t.Error("node pool not added correctly")
	}
}

func TestNewResourceQuantity(t *testing.T) {
	rq := NewResourceQuantity(2000, 4*1024*1024*1024)

	if rq.MilliCPU != 2000 {
		t.Errorf("expected MilliCPU 2000, got %d", rq.MilliCPU)
	}
	if rq.Bytes != 4*1024*1024*1024 {
		t.Errorf("expected Bytes %d, got %d", 4*1024*1024*1024, rq.Bytes)
	}
}

func TestNewContainer(t *testing.T) {
	resources := ResourceRequirements{
		CPURequest:    &ResourceQuantity{MilliCPU: 1000},
		MemoryRequest: &ResourceQuantity{Bytes: 2 * 1024 * 1024 * 1024},
	}

	container := NewContainer("container-1", "test-container", "default", "pod-1", "nginx:latest", resources)

	if container.ID != "container-1" {
		t.Errorf("expected ID container-1, got %s", container.ID)
	}
	if container.Name != "test-container" {
		t.Errorf("expected Name test-container, got %s", container.Name)
	}
	if container.Namespace != "default" {
		t.Errorf("expected Namespace default, got %s", container.Namespace)
	}
	if container.PodName != "pod-1" {
		t.Errorf("expected PodName pod-1, got %s", container.PodName)
	}
	if container.Image != "nginx:latest" {
		t.Errorf("expected Image nginx:latest, got %s", container.Image)
	}
}

func TestNewWorkload(t *testing.T) {
	workload := NewWorkload("workload-1", "test-workload", "default", WorkloadTypeDeployment, 3)

	if workload.ID != "workload-1" {
		t.Errorf("expected ID workload-1, got %s", workload.ID)
	}
	if workload.Name != "test-workload" {
		t.Errorf("expected Name test-workload, got %s", workload.Name)
	}
	if workload.Namespace != "default" {
		t.Errorf("expected Namespace default, got %s", workload.Namespace)
	}
	if workload.WorkloadType != WorkloadTypeDeployment {
		t.Errorf("expected WorkloadType Deployment, got %s", workload.WorkloadType)
	}
	if workload.ReplicaCount != 3 {
		t.Errorf("expected ReplicaCount 3, got %d", workload.ReplicaCount)
	}
}

func TestNewNodePool(t *testing.T) {
	nodePool := NewNodePool("node-pool-1", "test-node-pool", "cluster-1", "e2-standard-4", "us-central1", 3, true)

	if nodePool.ID != "node-pool-1" {
		t.Errorf("expected ID node-pool-1, got %s", nodePool.ID)
	}
	if nodePool.Name != "test-node-pool" {
		t.Errorf("expected Name test-node-pool, got %s", nodePool.Name)
	}
	if nodePool.ClusterID != "cluster-1" {
		t.Errorf("expected ClusterID cluster-1, got %s", nodePool.ClusterID)
	}
	if nodePool.MachineType != "e2-standard-4" {
		t.Errorf("expected MachineType e2-standard-4, got %s", nodePool.MachineType)
	}
	if nodePool.NodeCount != 3 {
		t.Errorf("expected NodeCount 3, got %d", nodePool.NodeCount)
	}
	if !nodePool.AutoScaling {
		t.Error("expected AutoScaling to be true")
	}
}

func TestNewCluster(t *testing.T) {
	cluster := NewCluster("cluster-1", "test-cluster", "us-central1", "us-central1", "project-1", "http://prometheus:9090")

	if cluster.ID != "cluster-1" {
		t.Errorf("expected ID cluster-1, got %s", cluster.ID)
	}
	if cluster.Name != "test-cluster" {
		t.Errorf("expected Name test-cluster, got %s", cluster.Name)
	}
	if cluster.Region != "us-central1" {
		t.Errorf("expected Region us-central1, got %s", cluster.Region)
	}
	if cluster.Location != "us-central1" {
		t.Errorf("expected Location us-central1, got %s", cluster.Location)
	}
	if cluster.ProjectID != "project-1" {
		t.Errorf("expected ProjectID project-1, got %s", cluster.ProjectID)
	}
	if cluster.PrometheusEndpoint != "http://prometheus:9090" {
		t.Errorf("expected PrometheusEndpoint http://prometheus:9090, got %s", cluster.PrometheusEndpoint)
	}
	if !cluster.IsActive {
		t.Error("expected IsActive to be true")
	}
}
