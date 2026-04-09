'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { getScoreLabel, formatCurrency, formatNumber } from '@/lib/utils';
import type { Cluster } from '@/types/api';
import { Server, MapPin, Database, Cpu, ArrowRight } from 'lucide-react';
import Link from 'next/link';

interface ClusterCardProps {
  cluster: Cluster;
}

export function ClusterCard({ cluster }: ClusterCardProps) {
  const scoreLabel = getScoreLabel(cluster.efficiencyScore ?? 0);
  const scoreColor =
    scoreLabel === 'GREEN'
      ? 'bg-green-500'
      : scoreLabel === 'YELLOW'
      ? 'bg-amber-500'
      : 'bg-red-500';

  return (
    <Card className="hover:shadow-lg transition-shadow">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center dark:bg-blue-900">
            <Server className="h-5 w-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <CardTitle className="text-lg">{cluster.name}</CardTitle>
            <div className="flex items-center gap-1 text-sm text-slate-500">
              <MapPin className="h-3 w-3" />
              {cluster.region}
            </div>
          </div>
        </div>
        <Badge className={cn('text-white', scoreColor)}>{scoreLabel}</Badge>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <p className="text-slate-500">Nodes</p>
            <p className="font-medium">{formatNumber(cluster.totalNodes)}</p>
          </div>
          <div>
            <p className="text-slate-500">Cost</p>
            <p className="font-medium">{formatCurrency(cluster.totalCost)}</p>
          </div>
          <div>
            <p className="text-slate-500">Score</p>
            <p className="font-medium">{cluster.efficiencyScore ?? 'N/A'}</p>
          </div>
        </div>
        <div className="mt-4 flex items-center justify-between border-t pt-4">
          <div className="flex gap-2">
            <div className="flex items-center gap-1 text-xs text-slate-500">
              <Cpu className="h-3 w-3" />
              {cluster.nodePools.length} node pools
            </div>
          </div>
          <Link href={`/efficiency?cluster=${cluster.id}`}>
            <Button variant="ghost" size="sm">
              View Details
              <ArrowRight className="ml-1 h-4 w-4" />
            </Button>
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}