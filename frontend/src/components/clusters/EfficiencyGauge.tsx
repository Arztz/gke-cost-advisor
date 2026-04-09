'use client';

import { cn, getScoreColor } from '@/lib/utils';

interface EfficiencyGaugeProps {
  score: number;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

export function EfficiencyGauge({ score, size = 'md', showLabel = true }: EfficiencyGaugeProps) {
  const color = getScoreColor(score);
  
  const sizes = {
    sm: { width: 60, strokeWidth: 6, fontSize: 'text-sm' },
    md: { width: 100, strokeWidth: 8, fontSize: 'text-lg' },
    lg: { width: 150, strokeWidth: 10, fontSize: 'text-2xl' },
  };
  
  const { width, strokeWidth, fontSize } = sizes[size];
  const radius = (width - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (score / 100) * circumference;
  
  return (
    <div className="flex flex-col items-center">
      <div className="relative" style={{ width, height: width }}>
        {/* Background circle */}
        <svg width={width} height={width} className="transform -rotate-90">
          <circle
            cx={width / 2}
            cy={width / 2}
            r={radius}
            fill="none"
            stroke="currentColor"
            strokeWidth={strokeWidth}
            className="text-slate-200 dark:text-slate-700"
          />
          <circle
            cx={width / 2}
            cy={width / 2}
            r={radius}
            fill="none"
            stroke={color}
            strokeWidth={strokeWidth}
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            className="transition-all duration-500"
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className={cn('font-bold', fontSize)} style={{ color }}>
            {score}
          </span>
        </div>
      </div>
      {showLabel && (
        <p className="mt-2 text-sm text-slate-500">Efficiency Score</p>
      )}
    </div>
  );
}