'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Separator } from '@/components/ui/separator';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { formatCurrency, formatDate, getScoreColor } from '@/lib/utils';
import type { Recommendation } from '@/types/api';
import {
  ChevronDown,
  ChevronUp,
  Copy,
  Terminal,
  ExternalLink,
  Zap,
  Server,
  Target,
  TrendingDown,
} from 'lucide-react';

interface RecommendationCardProps {
  recommendation: Recommendation;
}

const priorityColors = {
  high: 'bg-red-500',
  medium: 'bg-amber-500',
  low: 'bg-blue-500',
};

const typeIcons = {
  'right-sizing': Zap,
  'spot-migration': Server,
  'node-pool-optimization': Target,
};

export function RecommendationCard({ recommendation }: RecommendationCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [showKubectlDialog, setShowKubectlDialog] = useState(false);

  const TypeIcon = typeIcons[recommendation.type];

  return (
    <Card className="overflow-hidden">
      <CardHeader className="flex flex-row items-start justify-between pb-4">
        <div className="flex items-start gap-3">
          <div className={`mt-1 p-2 rounded-lg ${recommendation.type === 'right-sizing' ? 'bg-blue-100 dark:bg-blue-900' : recommendation.type === 'spot-migration' ? 'bg-purple-100 dark:bg-purple-900' : 'bg-green-100 dark:bg-green-900'}`}>
            <TypeIcon className={`h-5 w-5 ${recommendation.type === 'right-sizing' ? 'text-blue-600 dark:text-blue-400' : recommendation.type === 'spot-migration' ? 'text-purple-600 dark:text-purple-400' : 'text-green-600 dark:text-green-400'}`} />
          </div>
          <div>
            <CardTitle className="text-base">{recommendation.title}</CardTitle>
            <p className="text-sm text-slate-500 mt-1">{recommendation.namespace}</p>
          </div>
        </div>
        <Badge className={priorityColors[recommendation.priority]}>{recommendation.priority}</Badge>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-slate-600 dark:text-slate-300">{recommendation.description}</p>
        
        <div className="mt-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div>
              <p className="text-xs text-slate-500">Est. Savings</p>
              <p className="text-lg font-bold text-green-600">{formatCurrency(recommendation.estimatedSavings)}/mo</p>
            </div>
            <div>
              <p className="text-xs text-slate-500">Reduction</p>
              <p className="text-sm font-medium">{recommendation.savingsPercentage}%</p>
            </div>
          </div>
          <Button variant="outline" size="sm" onClick={() => setExpanded(!expanded)}>
            {expanded ? 'Less' : 'More'}
            {expanded ? <ChevronUp className="ml-1 h-4 w-4" /> : <ChevronDown className="ml-1 h-4 w-4" />}
          </Button>
        </div>

        {expanded && (
          <>
            <Separator className="my-4" />
            
            <div className="space-y-4">
              {/* Resource Changes */}
              <div>
                <h4 className="text-sm font-medium mb-2">Resource Changes</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="bg-slate-50 dark:bg-slate-900 p-3 rounded-lg">
                    <p className="text-slate-500 text-xs">Current</p>
                    <p className="font-medium">CPU: {recommendation.currentResources.cpu}</p>
                    <p className="font-medium">Memory: {recommendation.currentResources.memory}</p>
                  </div>
                  <div className="bg-green-50 dark:bg-green-900 p-3 rounded-lg">
                    <p className="text-green-600 text-xs">Recommended</p>
                    <p className="font-medium text-green-600">CPU: {recommendation.recommendedResources.cpu}</p>
                    <p className="font-medium text-green-600">Memory: {recommendation.recommendedResources.memory}</p>
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div>
                <h4 className="text-sm font-medium mb-2">Actions</h4>
                <div className="space-y-2">
                  {recommendation.actions.map((action, index) => (
                    <div key={index} className="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-900 rounded-lg">
                      <div>
                        <p className="text-sm font-medium">{action.description}</p>
                        <p className="text-xs text-slate-500 mt-1">{action.type}</p>
                      </div>
                      {action.type === 'kubectl' && (
                        <Button size="sm" variant="outline" onClick={() => setShowKubectlDialog(true)}>
                          <Terminal className="h-4 w-4 mr-1" />
                          View
                        </Button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex items-center justify-between text-xs text-slate-500">
                <span>Confidence: {recommendation.confidence}</span>
                <span>Expires: {formatDate(recommendation.expiresAt)}</span>
              </div>
            </div>
          </>
        )}

        {/* Kubectl Dialog */}
        <Dialog open={showKubectlDialog} onOpenChange={setShowKubectlDialog}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>kubectl Command</DialogTitle>
              <DialogDescription>Copy this command to apply the recommendation</DialogDescription>
            </DialogHeader>
            <div className="bg-slate-900 p-4 rounded-lg overflow-x-auto">
              <pre className="text-green-400 text-sm font-mono">{recommendation.actions.find(a => a.type === 'kubectl')?.command || 'N/A'}</pre>
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setShowKubectlDialog(false)}>Close</Button>
              <Button onClick={() => navigator.clipboard.writeText(recommendation.actions.find(a => a.type === 'kubectl')?.command || '')}>
                <Copy className="h-4 w-4 mr-2" />
                Copy
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}