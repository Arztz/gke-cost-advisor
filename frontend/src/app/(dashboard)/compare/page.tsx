'use client';

import { useState } from 'react';
import { useCompareMachineFamilies } from '@/lib/hooks/useApi';
import { RequirementsForm, RequirementsFormData } from '@/components/comparison/RequirementsForm';
import { ComparisonTable } from '@/components/comparison/ComparisonTable';
import { ComparisonChart } from '@/components/comparison/ComparisonChart';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Loader2, TrendingDown, Zap, DollarSign } from 'lucide-react';
import { formatCurrency } from '@/lib/utils';

export default function ComparePage() {
  const [showSpot, setShowSpot] = useState(true);
  const [formData, setFormData] = useState<RequirementsFormData | null>(null);
  
  const compareMutation = useCompareMachineFamilies();

  const handleSubmit = (data: RequirementsFormData) => {
    setFormData(data);
    compareMutation.mutate(data);
  };

  const result = compareMutation.data;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Compare Machine Families</h1>
          <p className="text-slate-500">Find the most cost-effective machine type for your workload</p>
        </div>
      </div>

      {/* Requirements Form */}
      <Card>
        <CardHeader>
          <CardTitle>Requirements</CardTitle>
        </CardHeader>
        <CardContent>
          <RequirementsForm
            onSubmit={handleSubmit}
            isLoading={compareMutation.isPending}
          />
        </CardContent>
      </Card>

      {/* Loading State */}
      {compareMutation.isPending && (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        </div>
      )}

      {/* Error State */}
      {compareMutation.isError && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="pt-4">
            <p className="text-red-600">Failed to compare machine families. Please try again.</p>
          </CardContent>
        </Card>
      )}

      {/* Results */}
      {result && !compareMutation.isPending && (
        <>
          {/* Summary Stats */}
          <div className="grid gap-4 md:grid-cols-4">
            <Card className="border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-green-700 dark:text-green-300">
                  Total Savings
                </CardTitle>
                <TrendingDown className="h-4 w-4 text-green-600 dark:text-green-400" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-700 dark:text-green-300">
                  {formatCurrency(result.totalSavings)}
                </div>
                <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                  per month
                </p>
              </CardContent>
            </Card>

            {showSpot && (
              <>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle className="text-sm font-medium text-slate-500">On-Demand Cost</CardTitle>
                    <DollarSign className="h-4 w-4 text-slate-400" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{formatCurrency(result.onDemandCost)}</div>
                    <p className="text-xs text-slate-500 mt-1">per month</p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle className="text-sm font-medium text-slate-500">Spot Cost</CardTitle>
                    <Zap className="h-4 w-4 text-slate-400" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{formatCurrency(result.spotCost)}</div>
                    <p className="text-xs text-slate-500 mt-1">per month</p>
                  </CardContent>
                </Card>

                <Card className="border-purple-200 bg-purple-50 dark:border-purple-800 dark:bg-purple-950">
                  <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle className="text-sm font-medium text-purple-700 dark:text-purple-300">
                      Spot Savings
                    </CardTitle>
                    <Zap className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-purple-700 dark:text-purple-300">
                      {result.savingsPercentage.toFixed(0)}%
                    </div>
                    <p className="text-xs text-purple-600 dark:text-purple-400 mt-1">off on-demand</p>
                  </CardContent>
                </Card>
              </>
            )}
          </div>

          {/* Toggle for spot */}
          <div className="flex items-center space-x-2">
            <Switch id="showSpot" checked={showSpot} onCheckedChange={setShowSpot} />
            <Label htmlFor="showSpot">Show Spot pricing</Label>
          </div>

          {/* Results Table and Chart */}
          <Tabs defaultValue="table">
            <TabsList>
              <TabsTrigger value="table">Table View</TabsTrigger>
              <TabsTrigger value="chart">Chart View</TabsTrigger>
            </TabsList>
            <TabsContent value="table">
              <ComparisonTable families={result.recommendedFamilies} showSpot={showSpot} />
            </TabsContent>
            <TabsContent value="chart">
              <Card>
                <CardHeader>
                  <CardTitle>Cost Comparison</CardTitle>
                </CardHeader>
                <CardContent>
                  <ComparisonChart families={result.recommendedFamilies} showSpot={showSpot} />
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>

          {/* Spot Savings Explanation */}
          {showSpot && (
            <Card className="bg-blue-50 border-blue-200 dark:bg-blue-950 dark:border-blue-800">
              <CardContent className="pt-4">
                <p className="text-sm text-blue-700 dark:text-blue-300">
                  <strong>Note:</strong> Spot instances can save 60-91% compared to on-demand pricing. 
                  However, they may be interrupted at any time. Use for fault-tolerant workloads or 
                  batch processing jobs.
                </p>
              </CardContent>
            </Card>
          )}
        </>
      )}

      {/* Empty State */}
      {!result && !compareMutation.isPending && !compareMutation.isError && (
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <p className="text-slate-500">Enter your requirements above to compare machine families</p>
        </div>
      )}
    </div>
  );
}