'use client';

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { cn, getScoreLabel, formatCurrency, formatNumber } from '@/lib/utils';
import type { Cluster } from '@/types/api';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

interface ClusterTableProps {
  clusters: Cluster[];
}

export function ClusterTable({ clusters }: ClusterTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Cluster Name</TableHead>
            <TableHead>Region</TableHead>
            <TableHead className="text-right">Nodes</TableHead>
            <TableHead className="text-right">Node Pools</TableHead>
            <TableHead className="text-right">Cost</TableHead>
            <TableHead className="text-center">Score</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {clusters.map((cluster) => {
            const scoreLabel = getScoreLabel(cluster.efficiencyScore ?? 0);
            const scoreColor =
              scoreLabel === 'GREEN'
                ? 'bg-green-500'
                : scoreLabel === 'YELLOW'
                ? 'bg-amber-500'
                : 'bg-red-500';

            return (
              <TableRow key={cluster.id}>
                <TableCell className="font-medium">{cluster.name}</TableCell>
                <TableCell>{cluster.region}</TableCell>
                <TableCell className="text-right">{formatNumber(cluster.totalNodes)}</TableCell>
                <TableCell className="text-right">{cluster.nodePools.length}</TableCell>
                <TableCell className="text-right">{formatCurrency(cluster.totalCost)}</TableCell>
                <TableCell className="text-center">
                  <Badge className={cn('text-white', scoreColor)}>
                    {cluster.efficiencyScore ?? 'N/A'}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <Link href={`/efficiency?cluster=${cluster.id}`}>
                    <Button variant="ghost" size="sm">
                      <ArrowRight className="h-4 w-4" />
                    </Button>
                  </Link>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}