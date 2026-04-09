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
import { formatCurrency } from '@/lib/utils';
import type { MachineFamily } from '@/types/api';
import { Star, Zap } from 'lucide-react';

interface ComparisonTableProps {
  families: MachineFamily[];
  showSpot?: boolean;
}

export function ComparisonTable({ families, showSpot = true }: ComparisonTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Machine Family</TableHead>
            <TableHead className="text-right">vCPUs</TableHead>
            <TableHead className="text-right">Memory (GB)</TableHead>
            <TableHead className="text-right">Storage (GB)</TableHead>
            <TableHead className="text-right">On-Demand</TableHead>
            {showSpot && <TableHead className="text-right">Spot</TableHead>}
            {showSpot && <TableHead className="text-center">Savings</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {families.map((family, index) => {
            const savings = ((family.onDemandPrice - family.spotPrice) / family.onDemandPrice) * 100;
            const isRecommended = index === 0;
            
            return (
              <TableRow key={family.name}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    {family.name}
                    {isRecommended && (
                      <Badge className="bg-green-500">
                        <Star className="h-3 w-3 mr-1" />
                        Best
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-right">{family.vCPUs}</TableCell>
                <TableCell className="text-right">{family.memoryGB}</TableCell>
                <TableCell className="text-right">{family.storageGB}</TableCell>
                <TableCell className="text-right">{formatCurrency(family.onDemandPrice)}/mo</TableCell>
                {showSpot && (
                  <>
                    <TableCell className="text-right">{formatCurrency(family.spotPrice)}/mo</TableCell>
                    <TableCell className="text-center">
                      <Badge className="bg-green-500">{savings.toFixed(0)}%</Badge>
                    </TableCell>
                  </>
                )}
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}