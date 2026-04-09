'use client';

import { useState, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { useClusters, useEfficiencyScores } from '@/lib/hooks/useApi';
import { MetricCard } from '@/components/efficiency/MetricCard';
import { UtilizationChart } from '@/components/efficiency/UtilizationChart';
import { PercentileTable } from '@/components/efficiency/PercentileTable';
import { TimeRangeSelector } from '@/components/efficiency/TimeRangeSelector';
import { EfficiencyGauge } from '@/components/clusters/EfficiencyGauge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Loader2, RefreshCw, AlertCircle, Clock, CheckCircle2 } from 'lucide-react';
import { formatDateTime } from '@/lib/utils';

function EfficiencyPageContent() {
  const searchParams = useSearchParams();
  const clusterIdFromUrl = searchParams.get('cluster');
  
  const { data: clusters, isLoading: clustersLoading } = useClusters();
  const [selectedClusterId, setSelectedClusterId] = useState<string>(clusterIdFromUrl || '');
  const [timeRange, setTimeRange] = useState('24h');
  
  const { data: efficiencyData, isLoading, isError, refetch } = useEfficiencyScores(
    selectedClusterId || (clusters?.[0]?.id ?? '')
  );

  if (clustersLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Efficiency Scores</h1>
          <p className="text-slate-500">Monitor resource utilization across namespaces</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <Select value={selectedClusterId} onValueChange={(value) => setSelectedClusterId(value || '')}>
          <SelectTrigger className="w-[250px]">
            <SelectValue placeholder="Select cluster" />
          </SelectTrigger>
          <SelectContent>
            {clusters?.map((cluster) => (
              <SelectItem key={cluster.id} value={cluster.id}>
                {cluster.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        
        <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
        
        <Button variant="outline" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        </div>
      ) : isError ? (
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <AlertCircle className="h-12 w-12 text-red-500" />
          <p className="text-slate-500">Failed to load efficiency data</p>
          <Button onClick={() => refetch()}>Retry</Button>
        </div>
      ) : efficiencyData ? (
        <>
          {/* Overall Score */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Cluster Efficiency</CardTitle>
                <p className="text-sm text-slate-500 mt-1">
                  {efficiencyData.clusterName}
                </p>
              </div>
              <EfficiencyGauge score={efficiencyData.overallScore} size="lg" />
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-4">
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-slate-400" />
                  <span className="text-sm text-slate-500">
                    Data: {efficiencyData.dataFreshness}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-slate-400" />
                  <span className="text-sm text-slate-500">
                    Confidence: {efficiencyData.confidence}
                  </span>
                </div>
                <div className="col-span-2 text-sm text-slate-500 text-right">
                  Last updated: {formatDateTime(efficiencyData.timestamp)}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Score Breakdown */}
          <div className="grid gap-4 md:grid-cols-3">
            <MetricCard
              title="CPU Efficiency"
              value={efficiencyData.cpuScore}
              score={efficiencyData.cpuScore}
            />
            <MetricCard
              title="Memory Efficiency"
              value={efficiencyData.memoryScore}
              score={efficiencyData.memoryScore}
            />
            <MetricCard
              title="Storage Efficiency"
              value={efficiencyData.storageScore}
              score={efficiencyData.storageScore}
            />
          </div>

          {/* Utilization Chart */}
          <Card>
            <CardHeader>
              <CardTitle>Utilization Over Time</CardTitle>
            </CardHeader>
            <CardContent>
              {/* Placeholder for actual time series data - in real app would come from API */}
              <UtilizationChart
                data={efficiencyData.namespaces.slice(0, 10).map((ns, i) => ({
                  timestamp: new Date(Date.now() - i * 3600000).toISOString(),
                  cpu: ns.cpuUtilization,
                  memory: ns.memoryUtilization,
                  storage: ns.storageUtilization,
                }))}
              />
            </CardContent>
          </Card>

          {/* Namespace Table */}
          <Card>
            <CardHeader>
              <CardTitle>Namespace Details</CardTitle>
            </CardHeader>
            <CardContent>
              <PercentileTable namespaces={efficiencyData.namespaces} />
            </CardContent>
          </Card>
        </>
      ) : (
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <p className="text-slate-500">Select a cluster to view efficiency data</p>
        </div>
      )}
    </div>
  );
}

export default function EfficiencyPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center h-96">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    }>
      <EfficiencyPageContent />
    </Suspense>
  );
}