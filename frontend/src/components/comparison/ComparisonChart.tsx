'use client';

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { formatCurrency } from '@/lib/utils';
import type { MachineFamily } from '@/types/api';

interface ComparisonChartProps {
  families: MachineFamily[];
  showSpot?: boolean;
}

export function ComparisonChart({ families, showSpot = true }: ComparisonChartProps) {
  const chartData = families.map((f) => ({
    name: f.name,
    'On-Demand': f.onDemandPrice,
    ...(showSpot ? { Spot: f.spotPrice } : {}),
  }));

  return (
    <div className="h-[350px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" className="stroke-slate-200" />
          <XAxis
            dataKey="name"
            tick={{ fontSize: 12, fill: '#64748B' }}
            tickLine={false}
            axisLine={{ stroke: '#E2E8F0' }}
          />
          <YAxis
            tick={{ fontSize: 12, fill: '#64748B' }}
            tickLine={false}
            axisLine={{ stroke: '#E2E8F0' }}
            tickFormatter={(value) => `$${value}`}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'white',
              border: '1px solid #E2E8F0',
              borderRadius: '8px',
              boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
            }}
            formatter={(value) => [value ? formatCurrency(Number(value)) : '', '']}
          />
          <Legend />
          <Bar dataKey="On-Demand" fill="#3B82F6" radius={[4, 4, 0, 0]} />
          {showSpot && <Bar dataKey="Spot" fill="#10B981" radius={[4, 4, 0, 0]} />}
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}