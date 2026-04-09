'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getClusters,
  getEfficiencyScores,
  getRecommendations,
  compareMachineFamilies,
  getHealth,
} from '@/lib/api';
import type { ComparisonRequest } from '@/types/api';

// Query keys
export const queryKeys = {
  clusters: ['clusters'] as const,
  cluster: (id: string) => ['clusters', id] as const,
  efficiency: (clusterId: string) => ['efficiency', clusterId] as const,
  recommendations: (filters?: Record<string, string>) => ['recommendations', filters] as const,
  comparison: (request: ComparisonRequest) => ['comparison', request] as const,
  health: ['health'] as const,
};

// Clusters hooks
export function useClusters() {
  return useQuery({
    queryKey: queryKeys.clusters,
    queryFn: getClusters,
    refetchInterval: 5 * 60 * 1000, // 5 minutes
    staleTime: 2 * 60 * 1000,
  });
}

export function useCluster(id: string) {
  return useQuery({
    queryKey: queryKeys.cluster(id),
    queryFn: () => getClusters().then((clusters) => clusters.find((c) => c.id === id)),
    enabled: !!id,
  });
}

// Efficiency hooks
export function useEfficiencyScores(clusterId: string) {
  return useQuery({
    queryKey: queryKeys.efficiency(clusterId),
    queryFn: () => getEfficiencyScores(clusterId),
    enabled: !!clusterId,
    staleTime: Infinity, // Manual refresh only
  });
}

// Recommendations hooks
export function useRecommendations(filters?: Record<string, string>) {
  return useQuery({
    queryKey: queryKeys.recommendations(filters),
    queryFn: () => getRecommendations(filters),
    refetchInterval: 1 * 60 * 1000, // 1 minute
    staleTime: 30 * 1000,
  });
}

// Comparison hook
export function useCompareMachineFamilies() {
  return useMutation({
    mutationFn: compareMachineFamilies,
  });
}

// Health hook
export function useHealth() {
  return useQuery({
    queryKey: queryKeys.health,
    queryFn: getHealth,
    refetchInterval: 30 * 1000,
  });
}