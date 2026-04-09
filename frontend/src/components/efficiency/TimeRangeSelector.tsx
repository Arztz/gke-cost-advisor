'use client';

import { Button } from '@/components/ui/button';

interface TimeRangeSelectorProps {
  value: string;
  onChange: (value: string) => void;
}

const timeRanges = [
  { label: '1 Hour', value: '1h' },
  { label: '6 Hours', value: '6h' },
  { label: '24 Hours', value: '24h' },
  { label: '7 Days', value: '7d' },
  { label: '30 Days', value: '30d' },
];

export function TimeRangeSelector({ value, onChange }: TimeRangeSelectorProps) {
  return (
    <div className="flex items-center gap-2">
      {timeRanges.map((range) => (
        <Button
          key={range.value}
          variant={value === range.value ? 'default' : 'outline'}
          size="sm"
          onClick={() => onChange(range.value)}
        >
          {range.label}
        </Button>
      ))}
    </div>
  );
}