'use client';

import { useClusters } from '@/lib/hooks/useApi';
import { ClusterCard } from '@/components/clusters/ClusterCard';
import { ClusterTable } from '@/components/clusters/ClusterTable';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, RefreshCw, AlertCircle } from 'lucide-react';

export default function ClustersPage() {
  const { data: clusters, isLoading, isError, refetch, error } = useClusters();

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
        <p className="text-slate-500">Failed to load clusters</p>
        <Button onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Retry
        </Button>
      </div>
    );
  }

  if (!clusters || clusters.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-96 gap-4">
        <p className="text-slate-500">No clusters found</p>
      </div>
    );
  }

  // Calculate summary stats
  const totalNodes = clusters.reduce((sum, c) => sum + c.totalNodes, 0);
  const totalCost = clusters.reduce((sum, c) => sum + c.totalCost, 0);
  const avgScore =
    clusters.reduce((sum, c) => sum + (c.efficiencyScore ?? 0), 0) / clusters.length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Clusters</h1>
          <p className="text-slate-500">Manage and monitor your GKE clusters</p>
        </div>
        <Button variant="outline" onClick={() => refetch()}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-slate-500">Total Clusters</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{clusters.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-slate-500">Total Nodes</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalNodes.toLocaleString()}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-slate-500">Avg Efficiency</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{avgScore.toFixed(0)}</div>
          </CardContent>
        </Card>
      </div>

      {/* Clusters List */}
      <Tabs defaultValue="grid">
        <TabsList>
          <TabsTrigger value="grid">Grid View</TabsTrigger>
          <TabsTrigger value="table">Table View</TabsTrigger>
        </TabsList>
        <TabsContent value="grid">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {clusters.map((cluster) => (
              <ClusterCard key={cluster.id} cluster={cluster} />
            ))}
          </div>
        </TabsContent>
        <TabsContent value="table">
          <ClusterTable clusters={clusters} />
        </TabsContent>
      </Tabs>
    </div>
  );
}