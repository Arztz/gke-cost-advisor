// Cluster types
export interface Cluster {
  id: string;
  name: string;
  region: string;
  nodePools: NodePool[];
  totalNodes: number;
  totalCost: number;
  efficiencyScore?: number;
  createdAt: string;
  updatedAt: string;
}

export interface NodePool {
  id: string;
  name: string;
  machineType: string;
  nodeCount: number;
  minNodes: number;
  maxNodes: number;
  autoScaling: boolean;
}

// Efficiency types
export interface EfficiencyScore {
  clusterId: string;
  clusterName: string;
  overallScore: number;
  cpuScore: number;
  memoryScore: number;
  storageScore: number;
  namespaces: NamespaceEfficiency[];
  timestamp: string;
  confidence: string;
  dataFreshness: string;
}

export interface NamespaceEfficiency {
  namespace: string;
  score: number;
  cpuUtilization: number;
  memoryUtilization: number;
  storageUtilization: number;
  podCount: number;
  requestCpu: number;
  limitCpu: number;
  requestMemory: number;
  limitMemory: number;
}

export interface UtilizationMetric {
  timestamp: string;
  cpu: number;
  memory: number;
  storage: number;
}

// Recommendation types
export interface Recommendation {
  id: string;
  type: 'right-sizing' | 'spot-migration' | 'node-pool-optimization';
  priority: 'high' | 'medium' | 'low';
  title: string;
  description: string;
  namespace: string;
  currentResources: ResourceSpec;
  recommendedResources: ResourceSpec;
  estimatedSavings: number;
  savingsPercentage: number;
  confidence: string;
  actions: RecommendationAction[];
  createdAt: string;
  expiresAt: string;
}

export interface ResourceSpec {
  cpu: string;
  memory: string;
  storage?: string;
  replicas?: number;
  instanceType?: string;
}

export interface RecommendationAction {
  type: 'kubectl' | 'gcloud' | 'console';
  command: string;
  description: string;
}

export interface RecommendationFilters {
  type?: string;
  priority?: string;
  namespace?: string;
  minSavings?: number;
}

// Machine family comparison types
export interface MachineFamily {
  name: string;
  vCPUs: number;
  memoryGB: number;
  storageGB: number;
  onDemandPrice: number;
  spotPrice: number;
  region: string;
}

export interface ComparisonRequest {
  vCPUs: number;
  memoryGB: number;
  storageGB: number;
  region: string;
  includeSpot: boolean;
}

export interface ComparisonResult {
  recommendedFamilies: MachineFamily[];
  totalSavings: number;
  savingsPercentage: number;
  spotSavings: number;
  onDemandCost: number;
  spotCost: number;
}

// Health check
export interface HealthStatus {
  status: 'healthy' | 'unhealthy' | 'degraded';
  checks: HealthCheck[];
  timestamp: string;
}

export interface HealthCheck {
  name: string;
  status: 'pass' | 'fail' | 'warn';
  message: string;
}

// API response wrapper
export interface ApiResponse<T> {
  data: T;
  error?: string;
  meta?: {
    timestamp: string;
    duration: number;
  };
}
