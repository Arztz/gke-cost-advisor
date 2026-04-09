'use client';

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { X } from 'lucide-react';

interface RecommendationFiltersProps {
  filters: {
    type: string;
    priority: string;
    namespace: string;
  };
  onFilterChange: (filters: Partial<RecommendationFiltersProps['filters']>) => void;
  onClearFilters: () => void;
  namespaces?: string[];
}

export function RecommendationFilters({
  filters,
  onFilterChange,
  onClearFilters,
  namespaces = [],
}: RecommendationFiltersProps) {
  const hasFilters = filters.type || filters.priority || filters.namespace;

  return (
    <div className="flex flex-wrap items-end gap-4">
      <div className="space-y-2">
        <Label className="text-xs text-slate-500">Type</Label>
        <Select value={filters.type} onValueChange={(value) => onFilterChange({ type: value || '' })}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="All types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All types</SelectItem>
            <SelectItem value="right-sizing">Right-sizing</SelectItem>
            <SelectItem value="spot-migration">Spot Migration</SelectItem>
            <SelectItem value="node-pool-optimization">Node Pool</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label className="text-xs text-slate-500">Priority</Label>
        <Select value={filters.priority} onValueChange={(value) => onFilterChange({ priority: value || '' })}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="All priorities" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All priorities</SelectItem>
            <SelectItem value="high">High</SelectItem>
            <SelectItem value="medium">Medium</SelectItem>
            <SelectItem value="low">Low</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label className="text-xs text-slate-500">Namespace</Label>
        <Select value={filters.namespace} onValueChange={(value) => onFilterChange({ namespace: value || '' })}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="All namespaces" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All namespaces</SelectItem>
            {namespaces.map((ns) => (
              <SelectItem key={ns} value={ns}>
                {ns}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {hasFilters && (
        <Button variant="ghost" size="sm" onClick={onClearFilters} className="text-slate-500">
          <X className="h-4 w-4 mr-1" />
          Clear
        </Button>
      )}
    </div>
  );
}