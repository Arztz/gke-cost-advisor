import axios from 'axios';
import type {
  Cluster,
  EfficiencyScore,
  Recommendation,
  ComparisonRequest,
  ComparisonResult,
  HealthStatus,
} from '@/types/api';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
});

// Clusters
export async function getClusters(): Promise<Cluster[]> {
  const response = await apiClient.get<Cluster[]>('/api/v1/clusters');
  return response.data;
}

export async function getCluster(id: string): Promise<Cluster> {
  const response = await apiClient.get<Cluster>(`/api/v1/clusters/${id}`);
  return response.data;
}

// Efficiency
export async function getEfficiencyScores(clusterId: string): Promise<EfficiencyScore> {
  const response = await apiClient.get<EfficiencyScore>(`/api/v1/clusters/${clusterId}/efficiency`);
  return response.data;
}

// Recommendations
export async function getRecommendations(filters?: Record<string, string>): Promise<Recommendation[]> {
  const response = await apiClient.get<Recommendation[]>('/api/v1/recommendations', {
    params: filters,
  });
  return response.data;
}

// Machine Family Comparison
export async function compareMachineFamilies(request: ComparisonRequest): Promise<ComparisonResult> {
  const response = await apiClient.post<ComparisonResult>('/api/v1/machine-families/compare', request);
  return response.data;
}

// Health
export async function getHealth(): Promise<HealthStatus> {
  const response = await apiClient.get<HealthStatus>('/health');
  return response.data;
}

export default apiClient;
