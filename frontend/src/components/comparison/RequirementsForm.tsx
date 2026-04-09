'use client';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { ArrowRight, Search } from 'lucide-react';

interface RequirementsFormProps {
  onSubmit: (data: RequirementsFormData) => void;
  isLoading?: boolean;
}

export interface RequirementsFormData {
  vCPUs: number;
  memoryGB: number;
  storageGB: number;
  region: string;
  includeSpot: boolean;
}

const regions = [
  'us-central1',
  'us-east1',
  'us-west1',
  'europe-west1',
  'europe-west2',
  'europe-west3',
  'asia-east1',
  'asia-northeast1',
  'asia-southeast1',
  'australia-southeast1',
];

export function RequirementsForm({ onSubmit, isLoading }: RequirementsFormProps) {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    
    onSubmit({
      vCPUs: Number(formData.get('vcpus')),
      memoryGB: Number(formData.get('memory')),
      storageGB: Number(formData.get('storage')),
      region: formData.get('region') as string,
      includeSpot: formData.get('includeSpot') === 'on',
    });
  };

  return (
    <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <div className="space-y-2">
        <Label htmlFor="vcpus">vCPUs</Label>
        <Input
          id="vcpus"
          name="vcpus"
          type="number"
          min="1"
          max="416"
          placeholder="4"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="memory">Memory (GB)</Label>
        <Input
          id="memory"
          name="memory"
          type="number"
          min="1"
          max="1408"
          placeholder="16"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="storage">Storage (GB)</Label>
        <Input
          id="storage"
          name="storage"
          type="number"
          min="0"
          max="64000"
          placeholder="100"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="region">Region</Label>
        <Select name="region" defaultValue="us-central1">
          <SelectTrigger>
            <SelectValue placeholder="Select region" />
          </SelectTrigger>
          <SelectContent>
            {regions.map((region) => (
              <SelectItem key={region} value={region}>
                {region}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex items-center space-x-2 pt-6">
        <Switch id="includeSpot" name="includeSpot" defaultChecked />
        <Label htmlFor="includeSpot">Include Spot pricing</Label>
      </div>

      <div className="flex items-end">
        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <span className="flex items-center">
              <svg className="animate-spin -ml-1 mr-2 h-4 w-4" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Comparing...
            </span>
          ) : (
            <>
              <Search className="h-4 w-4 mr-2" />
              Compare
            </>
          )}
        </Button>
      </div>
    </form>
  );
}