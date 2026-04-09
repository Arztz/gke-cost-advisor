'use client';

import Link from 'next/link';
import { useClusters, useRecommendations } from '@/lib/hooks/useApi';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { formatCurrency } from '@/lib/utils';
import {
  Server,
  Gauge,
  Lightbulb,
  Scale,
  TrendingDown,
  ArrowRight,
  Loader2,
  AlertCircle,
} from 'lucide-react';

export default function HomePage() {
  const { data: clusters, isLoading: clustersLoading } = useClusters();
  const { data: recommendations, isLoading: recommendationsLoading } = useRecommendations();

  const stats = {
    clusterCount: clusters?.length ?? 0,
    totalCost: clusters?.reduce((sum, c) => sum + c.totalCost, 0) ?? 0,
    totalNodes: clusters?.reduce((sum, c) => sum + c.totalNodes, 0) ?? 0,
    avgScore: clusters?.length
      ? clusters.reduce((sum, c) => sum + (c.efficiencyScore ?? 0), 0) / clusters.length
      : 0,
    recommendationCount: recommendations?.length ?? 0,
    potentialSavings: recommendations?.reduce((sum, r) => sum + r.estimatedSavings, 0) ?? 0,
  };

  const navItems = [
    {
      title: 'Clusters',
      description: 'View and manage your GKE clusters',
      href: '/clusters',
      icon: Server,
      color: 'bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-400',
    },
    {
      title: 'Efficiency',
      description: 'Monitor resource utilization scores',
      href: '/efficiency',
      icon: Gauge,
      color: 'bg-purple-100 text-purple-600 dark:bg-purple-900 dark:text-purple-400',
    },
    {
      title: 'Recommendations',
      description: 'View optimization recommendations',
      href: '/recommendations',
      icon: Lightbulb,
      color: 'bg-amber-100 text-amber-600 dark:bg-amber-900 dark:text-amber-400',
    },
    {
      title: 'Compare',
      description: 'Compare machine families',
      href: '/compare',
      icon: Scale,
      color: 'bg-green-100 text-green-600 dark:bg-green-900 dark:text-green-400',
    },
  ];

  if (clustersLoading || recommendationsLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white">Dashboard</h1>
          <p className="text-slate-500 mt-1">GKE Cost Optimization Overview</p>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-500">Clusters</CardTitle>
            <Server className="h-4 w-4 text-slate-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.clusterCount}</div>
            <p className="text-xs text-slate-500 mt-1">{stats.totalNodes.toLocaleString()} total nodes</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-500">Monthly Cost</CardTitle>
            <TrendingDown className="h-4 w-4 text-slate-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(stats.totalCost)}</div>
            <p className="text-xs text-slate-500 mt-1">current spend</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-slate-500">Avg Efficiency</CardTitle>
            <Gauge className="h-4 w-4 text-slate-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.avgScore.toFixed(0)}</div>
            <p className="text-xs text-slate-500 mt-1">overall score</p>
          </CardContent>
        </Card>

        <Card className="border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-green-700 dark:text-green-300">
              Potential Savings
            </CardTitle>
            <TrendingDown className="h-4 w-4 text-green-600 dark:text-green-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-700 dark:text-green-300">
              {formatCurrency(stats.potentialSavings)}
            </div>
            <p className="text-xs text-green-600 dark:text-green-400 mt-1">
              {stats.recommendationCount} recommendations
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Navigation Cards */}
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white mb-4">Quick Navigation</h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {navItems.map((item) => (
            <Link key={item.href} href={item.href}>
              <Card className="hover:shadow-lg transition-all cursor-pointer group h-full">
                <CardContent className="pt-6">
                  <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${item.color} mb-4`}>
                    <item.icon className="h-6 w-6" />
                  </div>
                  <h3 className="font-semibold text-slate-900 dark:text-white group-hover:text-blue-600 transition-colors">
                    {item.title}
                  </h3>
                  <p className="text-sm text-slate-500 mt-1">{item.description}</p>
                  <div className="mt-4 flex items-center text-blue-600 text-sm font-medium">
                    Go to {item.title.toLowerCase()}
                    <ArrowRight className="h-4 w-4 ml-1 group-hover:translate-x-1 transition-transform" />
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </div>

      {/* Recent Recommendations Preview */}
      {recommendations && recommendations.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Recent Recommendations</CardTitle>
            <Link href="/recommendations">
              <Button variant="ghost" size="sm">
                View All
                <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {recommendations.slice(0, 3).map((rec) => (
                <div
                  key={rec.id}
                  className="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-900 rounded-lg"
                >
                  <div>
                    <p className="font-medium text-slate-900 dark:text-white">{rec.title}</p>
                    <p className="text-sm text-slate-500">{rec.namespace}</p>
                  </div>
                  <div className="text-right">
                    <p className="font-semibold text-green-600">{formatCurrency(rec.estimatedSavings)}/mo</p>
                    <p className="text-xs text-slate-500">{rec.savingsPercentage}% savings</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}