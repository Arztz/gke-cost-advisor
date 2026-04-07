package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gke-cost-advisor/internal/domain/entities"
	"gke-cost-advisor/internal/domain/valueobjects"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// PrometheusClient handles communication with Prometheus
type PrometheusClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewPrometheusClient creates a new Prometheus client
func NewPrometheusClient(endpoint, token string) *PrometheusClient {
	return &PrometheusClient{
		baseURL: endpoint,
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Query executes a PromQL query
func (c *PrometheusClient) Query(ctx context.Context, query string, t time.Time) ([]Sample, error) {
	url := fmt.Sprintf("%s/api/v1/query", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result prometheusQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	samples := make([]Sample, 0, len(result.Data.Result))
	for _, r := range result.Data.Result {
		var val float64
		switch v := r.Value[1].(type) {
		case float64:
			val = v
		case string:
			fmt.Sscanf(v, "%f", &val)
		}
		samples = append(samples, Sample{
			Metric: r.Metric,
			Value:  val,
		})
	}

	return samples, nil
}

// QueryRange executes a PromQL range query
func (c *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]Series, error) {
	url := fmt.Sprintf("%s/api/v1/query_range", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", start.Format(time.RFC3339))
	q.Add("end", end.Format(time.RFC3339))
	q.Add("step", step.String())
	req.URL.RawQuery = q.Encode()

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result prometheusRangeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	series := make([]Series, 0, len(result.Data.Result))
	for _, r := range result.Data.Result {
		series = append(series, Series{
			Metric: r.Metric,
			Values: r.Values,
		})
	}

	return series, nil
}

// GetContainerCPUUtilization gets CPU utilization for a container
func (c *PrometheusClient) GetContainerCPUUtilization(ctx context.Context, namespace, pod, container string, start, end time.Time, step time.Duration) ([]Series, error) {
	query := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{namespace="%s", pod="%s", container="%s"}[5m])`, namespace, pod, container)
	return c.QueryRange(ctx, query, start, end, step)
}

// GetContainerMemoryWorkingSet gets memory working set for a container
func (c *PrometheusClient) GetContainerMemoryWorkingSet(ctx context.Context, namespace, pod, container string, start, end time.Time, step time.Duration) ([]Series, error) {
	query := fmt.Sprintf(`container_memory_working_set_bytes{namespace="%s", pod="%s", container="%s"}`, namespace, pod, container)
	return c.QueryRange(ctx, query, start, end, step)
}

// GetContainerThrottling gets throttling metrics for a container
func (c *PrometheusClient) GetContainerThrottling(ctx context.Context, namespace, pod, container string, start, end time.Time, step time.Duration) ([]Series, error) {
	query := fmt.Sprintf(`rate(container_cpu_cfs_throttled_seconds_total{namespace="%s", pod="%s", container="%s"}[5m])`, namespace, pod, container)
	return c.QueryRange(ctx, query, start, end, step)
}

// GetNamespaceCPUUtilization gets CPU utilization for a namespace
func (c *PrometheusClient) GetNamespaceCPUUtilization(ctx context.Context, namespace string, start, end time.Time, step time.Duration) ([]Series, error) {
	query := fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s"}[5m])) by (namespace)`, namespace)
	return c.QueryRange(ctx, query, start, end, step)
}

// GetNamespaceMemoryUtilization gets memory utilization for a namespace
func (c *PrometheusClient) GetNamespaceMemoryUtilization(ctx context.Context, namespace string, start, end time.Time, step time.Duration) ([]Series, error) {
	query := fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s"}) by (namespace)`, namespace)
	return c.QueryRange(ctx, query, start, end, step)
}

// CalculatePercentiles calculates percentile values from time series data
func (c *PrometheusClient) CalculatePercentiles(series []Series) valueobjects.PercentileData {
	if len(series) == 0 || len(series[0].Values) == 0 {
		return valueobjects.PercentileData{}
	}

	values := make([]float64, 0, len(series[0].Values))
	for _, v := range series[0].Values {
		if len(v) >= 2 {
			switch val := v[1].(type) {
			case float64:
				values = append(values, val)
			case string:
				var f float64
				fmt.Sscanf(val, "%f", &f)
				values = append(values, f)
			}
		}
	}

	if len(values) == 0 {
		return valueobjects.PercentileData{}
	}

	// Sort values for percentile calculation
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	calcPercentile := func(p float64) float64 {
		idx := int(float64(len(values)-1) * p)
		if idx >= len(values) {
			idx = len(values) - 1
		}
		return values[idx]
	}

	return valueobjects.PercentileData{
		P50CPU:    calcPercentile(0.50),
		P90CPU:    calcPercentile(0.90),
		P95CPU:    calcPercentile(0.95),
		P99CPU:    calcPercentile(0.99),
		P50Memory: calcPercentile(0.50),
		P90Memory: calcPercentile(0.90),
		P95Memory: calcPercentile(0.95),
		P99Memory: calcPercentile(0.99),
	}
}

// Sample represents a Prometheus sample
type Sample struct {
	Metric map[string]string
	Value  float64
}

// Series represents a Prometheus time series
type Series struct {
	Metric map[string]string
	Values [][]interface{}
}

type prometheusQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type prometheusRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// KubernetesClient handles communication with Kubernetes API
type KubernetesClient struct {
	client         kubernetes.Interface
	dynamicClient  dynamic.Interface
	config         *rest.Config
	kubeconfigPath string
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient(kubeconfigPath string) (*KubernetesClient, error) {
	var cfg *rest.Config
	var err error

	if kubeconfigPath != "" {
		// Try in-cluster first, then load from file
		cfg, err = rest.InClusterConfig()
		if err != nil {
			cfg, err = loadKubeconfig(kubeconfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
			}
		}
	} else {
		// Try in-cluster config first, then default kubeconfig
		cfg, err = rest.InClusterConfig()
		if err != nil {
			cfg, err = loadKubeconfig(os.Getenv("HOME") + "/.kube/config")
			if err != nil {
				return nil, fmt.Errorf("no kubeconfig available: %w", err)
			}
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &KubernetesClient{
		client:         clientset,
		dynamicClient:  dynClient,
		config:         cfg,
		kubeconfigPath: kubeconfigPath,
	}, nil
}

// loadKubeconfig loads kubeconfig from file
func loadKubeconfig(path string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", path)
}

// ListNamespaces lists all namespaces
func (c *KubernetesClient) ListNamespaces(ctx context.Context) ([]string, error) {
	nsList, err := c.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

// ListPods lists pods in a namespace
func (c *KubernetesClient) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	podInfos := make([]PodInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podInfos = append(podInfos, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
		})
	}

	return podInfos, nil
}

// GetPodContainers gets containers for a specific pod
func (c *KubernetesClient) GetPodContainers(ctx context.Context, namespace, podName string) ([]ContainerInfo, error) {
	pod, err := c.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]ContainerInfo, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		var cpuReq, memReq int64
		var cpuLim, memLim int64

		if container.Resources.Requests != nil {
			cpuReq = container.Resources.Requests.Cpu().MilliValue()
			memReq = container.Resources.Requests.Memory().Value()
		}
		if container.Resources.Limits != nil {
			cpuLim = container.Resources.Limits.Cpu().MilliValue()
			memLim = container.Resources.Limits.Memory().Value()
		}

		containers = append(containers, ContainerInfo{
			Name:          container.Name,
			Image:         container.Image,
			CPURequest:    cpuReq,
			MemoryRequest: memReq,
			CPULimit:      cpuLim,
			MemoryLimit:   memLim,
		})
	}

	return containers, nil
}

// GetAllPodsInNamespace gets all pods with their resource requirements
func (c *KubernetesClient) GetAllPodsInNamespace(ctx context.Context, namespace string) ([]PodDetails, error) {
	pods, err := c.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	podDetails := make([]PodDetails, 0, len(pods.Items))
	for _, pod := range pods.Items {
		containers := make([]ContainerInfo, 0, len(pod.Spec.Containers))
		for _, container := range pod.Spec.Containers {
			var cpuReq, memReq int64
			var cpuLim, memLim int64

			if container.Resources.Requests != nil {
				if container.Resources.Requests.Cpu() != nil {
					cpuReq = container.Resources.Requests.Cpu().MilliValue()
				}
				if container.Resources.Requests.Memory() != nil {
					memReq = container.Resources.Requests.Memory().Value()
				}
			}
			if container.Resources.Limits != nil {
				if container.Resources.Limits.Cpu() != nil {
					cpuLim = container.Resources.Limits.Cpu().MilliValue()
				}
				if container.Resources.Limits.Memory() != nil {
					memLim = container.Resources.Limits.Memory().Value()
				}
			}

			containers = append(containers, ContainerInfo{
				Name:          container.Name,
				Image:         container.Image,
				CPURequest:    cpuReq,
				MemoryRequest: memReq,
				CPULimit:      cpuLim,
				MemoryLimit:   memLim,
			})
		}

		podDetails = append(podDetails, PodDetails{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Labels:     pod.Labels,
			Containers: containers,
		})
	}

	return podDetails, nil
}

// GetNodePools gets node pools (GKE-specific)
func (c *KubernetesClient) GetNodePools(ctx context.Context, cluster string) ([]*entities.NodePool, error) {
	return c.inferNodePoolsFromNodes(ctx, cluster)
}

// inferNodePoolsFromNodes infers node pools from actual nodes
func (c *KubernetesClient) inferNodePoolsFromNodes(ctx context.Context, cluster string) ([]*entities.NodePool, error) {
	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodePoolsMap := make(map[string]*entities.NodePool)
	for _, node := range nodes.Items {
		labels := node.Labels
		poolName := labels["cloud.google.com/gke-nodepool"]
		if poolName == "" {
			poolName = "default-pool"
		}

		if np, ok := nodePoolsMap[poolName]; ok {
			np.NodeCount++
		} else {
			nodePoolsMap[poolName] = &entities.NodePool{
				ID:          poolName,
				Name:        poolName,
				ClusterID:   cluster,
				NodeCount:   1,
				MachineType: labels["node.kubernetes.io/instance-type"],
			}
		}
	}

	nodePools := make([]*entities.NodePool, 0, len(nodePoolsMap))
	for _, np := range nodePoolsMap {
		nodePools = append(nodePools, np)
	}

	return nodePools, nil
}

// PodInfo represents simplified pod information
type PodInfo struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

// PodDetails represents detailed pod information with containers
type PodDetails struct {
	Name       string
	Namespace  string
	Labels     map[string]string
	Containers []ContainerInfo
}

// ContainerInfo represents container information
type ContainerInfo struct {
	Name          string
	Image         string
	CPURequest    int64
	MemoryRequest int64
	CPULimit      int64
	MemoryLimit   int64
}

// Pod represents a Kubernetes pod (simplified)
type Pod struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

// BillingClient handles communication with Google Cloud Billing API
type BillingClient struct {
	projectID        string
	billingAccountID string
	httpClient       *http.Client
}

// NewBillingClient creates a new Billing client
func NewBillingClient(projectID, billingAccountID string) *BillingClient {
	return &BillingClient{
		projectID:        projectID,
		billingAccountID: billingAccountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetMachineTypePricing gets pricing for a machine type
func (c *BillingClient) GetMachineTypePricing(ctx context.Context, machineType, region string) (*valueobjects.PricingInfo, error) {
	return c.getStaticPricing(machineType, region)
}

// getStaticPricing returns static pricing data for common machine types
func (c *BillingClient) getStaticPricing(machineType, region string) (*valueobjects.PricingInfo, error) {
	machineFamily := getMachineFamily(machineType)

	// Static pricing data for common machine types (USD per hour)
	staticPricing := map[string]map[string]struct {
		OnDemand float64
		Spot     float64
	}{
		"us-central1": {
			"e2-standard-2": {OnDemand: 0.096, Spot: 0.029},
			"e2-standard-4": {OnDemand: 0.192, Spot: 0.058},
			"e2-standard-8": {OnDemand: 0.384, Spot: 0.115},
			"e2-highmem-2":  {OnDemand: 0.153, Spot: 0.046},
			"e2-highmem-4":  {OnDemand: 0.306, Spot: 0.092},
			"e2-highcpu-2":  {OnDemand: 0.084, Spot: 0.025},
			"e2-highcpu-4":  {OnDemand: 0.168, Spot: 0.050},
			"n2-standard-2": {OnDemand: 0.112, Spot: 0.034},
			"n2-standard-4": {OnDemand: 0.224, Spot: 0.067},
			"n2-standard-8": {OnDemand: 0.448, Spot: 0.134},
			"n2-highmem-2":  {OnDemand: 0.170, Spot: 0.051},
			"n2-highmem-4":  {OnDemand: 0.340, Spot: 0.102},
			"n2-highcpu-2":  {OnDemand: 0.096, Spot: 0.029},
			"n2-highcpu-4":  {OnDemand: 0.192, Spot: 0.058},
			"c3-standard-2": {OnDemand: 0.146, Spot: 0.044},
			"c3-standard-4": {OnDemand: 0.291, Spot: 0.087},
			"c3-standard-8": {OnDemand: 0.582, Spot: 0.175},
			"c3-highmem-2":  {OnDemand: 0.228, Spot: 0.068},
			"c3-highmem-4":  {OnDemand: 0.456, Spot: 0.137},
			"c3-highcpu-2":  {OnDemand: 0.124, Spot: 0.037},
			"c3-highcpu-4":  {OnDemand: 0.248, Spot: 0.074},
		},
		"us-east1": {
			"e2-standard-4": {OnDemand: 0.192, Spot: 0.058},
			"n2-standard-4": {OnDemand: 0.224, Spot: 0.067},
			"c3-standard-4": {OnDemand: 0.291, Spot: 0.087},
		},
		"europe-west1": {
			"e2-standard-4": {OnDemand: 0.210, Spot: 0.063},
			"n2-standard-4": {OnDemand: 0.252, Spot: 0.076},
			"c3-standard-4": {OnDemand: 0.324, Spot: 0.097},
		},
		"asia-northeast1": {
			"e2-standard-4": {OnDemand: 0.230, Spot: 0.069},
			"n2-standard-4": {OnDemand: 0.266, Spot: 0.080},
			"c3-standard-4": {OnDemand: 0.345, Spot: 0.104},
		},
	}

	regionPricing, ok := staticPricing[region]
	if !ok {
		regionPricing = staticPricing["us-central1"]
	}

	pricing, ok := regionPricing[machineType]
	if !ok {
		// Default pricing for unknown types
		return valueobjects.NewPricingInfo(machineType, machineFamily, region, 0.192, 0.057), nil
	}

	return valueobjects.NewPricingInfo(machineType, machineFamily, region, pricing.OnDemand, pricing.Spot), nil
}

// GetAllMachineTypePricing gets all machine type pricing for a region
func (c *BillingClient) GetAllMachineTypePricing(ctx context.Context, region string) (map[string]*valueobjects.PricingInfo, error) {
	machineTypes := []string{
		"e2-standard-2", "e2-standard-4", "e2-standard-8",
		"e2-highmem-2", "e2-highmem-4",
		"e2-highcpu-2", "e2-highcpu-4",
		"n2-standard-2", "n2-standard-4", "n2-standard-8",
		"n2-highmem-2", "n2-highmem-4",
		"n2-highcpu-2", "n2-highcpu-4",
		"c3-standard-2", "c3-standard-4", "c3-standard-8",
		"c3-highmem-2", "c3-highmem-4",
		"c3-highcpu-2", "c3-highcpu-4",
	}

	pricing := make(map[string]*valueobjects.PricingInfo, len(machineTypes))
	for _, mt := range machineTypes {
		p, err := c.GetMachineTypePricing(ctx, mt, region)
		if err == nil {
			pricing[mt] = p
		}
	}

	return pricing, nil
}

// getMachineFamily extracts machine family from machine type
func getMachineFamily(machineType string) string {
	families := []string{"e2", "n2", "n1", "c2", "c3", "m1", "m2", "a2"}
	for _, family := range families {
		if len(machineType) >= len(family) && machineType[:len(family)] == family {
			return family
		}
	}
	return "unknown"
}
