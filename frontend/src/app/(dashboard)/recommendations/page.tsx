'use client';

import { useState, useMemo } from 'react';
import { useRecommendations } from '@/lib/hooks/useApi';
import { RecommendationCard } from '@/components/recommendations/RecommendationCard';
import { RecommendationFilters } from '@/components/recommendations/RecommendationFilters';
import { SavingsSummary } from '@/components/recommendations/SavingsSummary';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Loader2, RefreshCw, AlertCircle, Download, FileJson, FileText } from 'lucide-react';
import type { Recommendation } from '@/types/api';

function exportToCSV(recommendations: Recommendation[]) {
  const headers = ['Type', 'Priority', 'Title', 'Namespace', 'Current CPU', 'Recommended CPU', 'Current Memory', 'Recommended Memory', 'Savings', 'Savings %'];
  const rows = recommendations.map(r => [
    r.type,
    r.priority,
    `"${r.title}"`,
    r.namespace,
    r.currentResources.cpu,
    r.recommendedResources.cpu,
    r.currentResources.memory,
    r.recommendedResources.memory,
    r.estimatedSavings,
    r.savingsPercentage,
  ]);
  
  const csv = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
  const blob = new Blob([csv], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'recommendations.csv';
  a.click();
  URL.revokeObjectURL(url);
}

function exportToJSON(recommendations: Recommendation[]) {
  const json = JSON.stringify(recommendations, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'recommendations.json';
  a.click();
  URL.revokeObjectURL(url);
}

export default function RecommendationsPage() {
  const [filters, setFilters] = useState({
    type: '',
    priority: '',
    namespace: '',
  });
  
  const { data: recommendations, isLoading, isError, refetch } = useRecommendations();

  const filteredRecommendations = useMemo(() => {
    if (!recommendations) return [];
    
    return recommendations.filter((r) => {
      if (filters.type && filters.type !== 'all' && r.type !== filters.type) return false;
      if (filters.priority && filters.priority !== 'all' && r.priority !== filters.priority) return false;
      if (filters.namespace && filters.namespace !== 'all' && r.namespace !== filters.namespace) return false;
      return true;
    });
  }, [recommendations, filters]);

  const summaryStats = useMemo(() => {
    if (!filteredRecommendations.length) {
      return { totalSavings: 0, count: 0, avgPercentage: 0 };
    }
    
    const totalSavings = filteredRecommendations.reduce((sum, r) => sum + r.estimatedSavings, 0);
    const avgPercentage = filteredRecommendations.reduce((sum, r) => sum + r.savingsPercentage, 0) / filteredRecommendations.length;
    
    return {
      totalSavings,
      count: filteredRecommendations.length,
      avgPercentage,
    };
  }, [filteredRecommendations]);

  // Get unique namespaces for filter
  const namespaces = useMemo(() => {
    if (!recommendations) return [];
    return [...new Set(recommendations.map(r => r.namespace))];
  }, [recommendations]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center h-96 gap-4">
        <AlertCircle className="h-12 w-12 text-red-500" />
        <p className="text-slate-500">Failed to load recommendations</p>
        <Button onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Recommendations</h1>
          <p className="text-slate-500">Optimization recommendations for your clusters</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button variant="outline" onClick={() => exportToCSV(filteredRecommendations)}>
            <FileText className="h-4 w-4 mr-2" />
            CSV
          </Button>
          <Button variant="outline" onClick={() => exportToJSON(filteredRecommendations)}>
            <FileJson className="h-4 w-4 mr-2" />
            JSON
          </Button>
        </div>
      </div>

      {/* Savings Summary */}
      {recommendations && recommendations.length > 0 && (
        <SavingsSummary
          totalSavings={summaryStats.totalSavings}
          recommendationCount={summaryStats.count}
          avgSavingsPercentage={summaryStats.avgPercentage}
        />
      )}

      {/* Filters */}
      <Card>
        <CardContent className="pt-4">
          <RecommendationFilters
            filters={filters}
            onFilterChange={(newFilters) => setFilters(prev => ({ ...prev, ...newFilters }))}
            onClearFilters={() => setFilters({ type: '', priority: '', namespace: '' })}
            namespaces={namespaces}
          />
        </CardContent>
      </Card>

      {/* Recommendations List */}
      {filteredRecommendations.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <p className="text-slate-500">No recommendations found</p>
        </div>
      ) : (
        <div className="grid gap-4">
          {filteredRecommendations.map((rec) => (
            <RecommendationCard key={rec.id} recommendation={rec} />
          ))}
        </div>
      )}
    </div>
  );
}