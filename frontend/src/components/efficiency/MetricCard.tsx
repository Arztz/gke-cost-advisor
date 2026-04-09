'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn, getScoreColor, formatPercent } from '@/lib/utils';

interface MetricCardProps {
  title: string;
  value: number;
  unit?: string;
  score?: number;
  icon?: React.ReactNode;
}

export function MetricCard({ title, value, unit = '%', score, icon }: MetricCardProps) {
  const scoreColor = score ? getScoreColor(score) : undefined;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-slate-500">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-1">
          <div className="text-2xl font-bold" style={scoreColor ? { color: scoreColor } : undefined}>
            {value.toFixed(1)}
          </div>
          <span className="text-sm text-slate-500">{unit}</span>
        </div>
        {score !== undefined && (
          <p className="mt-1 text-xs text-slate-500">Score: {score}</p>
        )}
      </CardContent>
    </Card>
  );
}