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
import { getScoreColor } from '@/lib/utils';
import type { NamespaceEfficiency } from '@/types/api';

interface PercentileTableProps {
  namespaces: NamespaceEfficiency[];
}

export function PercentileTable({ namespaces }: PercentileTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Namespace</TableHead>
            <TableHead className="text-right">Pods</TableHead>
            <TableHead className="text-right">CPU Util</TableHead>
            <TableHead className="text-right">Memory Util</TableHead>
            <TableHead className="text-right">Storage Util</TableHead>
            <TableHead className="text-center">Score</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {namespaces.map((ns) => {
            const scoreColor = getScoreColor(ns.score);
            return (
              <TableRow key={ns.namespace}>
                <TableCell className="font-medium">{ns.namespace}</TableCell>
                <TableCell className="text-right">{ns.podCount}</TableCell>
                <TableCell className="text-right">{ns.cpuUtilization.toFixed(1)}%</TableCell>
                <TableCell className="text-right">{ns.memoryUtilization.toFixed(1)}%</TableCell>
                <TableCell className="text-right">{ns.storageUtilization.toFixed(1)}%</TableCell>
                <TableCell className="text-center">
                  <Badge
                    style={{
                      backgroundColor: scoreColor,
                      color: 'white',
                    }}
                  >
                    {ns.score}
                  </Badge>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}