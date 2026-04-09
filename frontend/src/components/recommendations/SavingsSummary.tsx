'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { formatCurrency } from '@/lib/utils';
import { TrendingDown, DollarSign, Percent } from 'lucide-react';

interface SavingsSummaryProps {
  totalSavings: number;
  recommendationCount: number;
  avgSavingsPercentage: number;
}

export function SavingsSummary({ totalSavings, recommendationCount, avgSavingsPercentage }: SavingsSummaryProps) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Card className="border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium text-green-700 dark:text-green-300">
            Potential Monthly Savings
          </CardTitle>
          <TrendingDown className="h-4 w-4 text-green-600 dark:text-green-400" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-green-700 dark:text-green-300">
            {formatCurrency(totalSavings)}
          </div>
          <p className="text-xs text-green-600 dark:text-green-400 mt-1">
            across {recommendationCount} recommendations
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium text-slate-500">Avg. Reduction</CardTitle>
          <Percent className="h-4 w-4 text-slate-400" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{avgSavingsPercentage.toFixed(0)}%</div>
          <p className="text-xs text-slate-500 mt-1">average savings per recommendation</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium text-slate-500">Annual Projection</CardTitle>
          <DollarSign className="h-4 w-4 text-slate-400" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{formatCurrency(totalSavings * 12)}</div>
          <p className="text-xs text-slate-500 mt-1">if all recommendations applied</p>
        </CardContent>
      </Card>
    </div>
  );
}