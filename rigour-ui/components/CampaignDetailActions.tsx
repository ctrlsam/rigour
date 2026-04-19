'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Play, Pause, Square, Trash2, RotateCcw, Loader2 } from 'lucide-react';
import { Campaign, CampaignStatus } from '../lib/types';
import * as api from '../lib/api';
import { Button } from './ui/button';

export function CampaignDetailActions({ campaign }: { campaign: Campaign }) {
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const [optimisticStatus, setOptimisticStatus] = useState<CampaignStatus | null>(null);

  const status = optimisticStatus ?? campaign.status;

  async function run(action: () => Promise<void>, nextStatus: CampaignStatus | null, redirectOnDelete = false) {
    setError('');
    setPending(true);
    if (nextStatus) setOptimisticStatus(nextStatus);
    try {
      await action();
      if (redirectOnDelete) {
        router.push('/campaigns');
      } else {
        router.refresh();
      }
    } catch (e) {
      setOptimisticStatus(null);
      setError(e instanceof Error ? e.message : 'Action failed');
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="flex flex-col items-end gap-2">
      <div className="flex gap-2 flex-wrap justify-end">
        {status === 'pending' && (
          <Button
            size="sm"
            onClick={() => run(() => api.startCampaign(campaign.id), 'running')}
            disabled={pending}
            className="gap-1.5"
          >
            {pending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Play className="h-3 w-3" />}
            Start
          </Button>
        )}
        {status === 'running' && (
          <Button
            size="sm"
            variant="outline"
            onClick={() => run(() => api.pauseCampaign(campaign.id), 'paused')}
            disabled={pending}
            className="gap-1.5"
          >
            {pending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Pause className="h-3 w-3" />}
            Pause
          </Button>
        )}
        {status === 'paused' && (
          <Button
            size="sm"
            onClick={() => run(() => api.resumeCampaign(campaign.id), 'running')}
            disabled={pending}
            className="gap-1.5"
          >
            {pending ? <Loader2 className="h-3 w-3 animate-spin" /> : <RotateCcw className="h-3 w-3" />}
            Resume
          </Button>
        )}
        {(status === 'running' || status === 'paused') && (
          <Button
            size="sm"
            variant="outline"
            onClick={() => run(() => api.cancelCampaign(campaign.id), 'cancelled')}
            disabled={pending}
            className="gap-1.5 text-destructive border-destructive/30 hover:bg-destructive/10"
          >
            {pending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Square className="h-3 w-3" />}
            Cancel
          </Button>
        )}
        {(status === 'completed' ||
          status === 'failed' ||
          status === 'cancelled' ||
          status === 'pending') && (
          <Button
            size="sm"
            variant="ghost"
            onClick={() => run(() => api.deleteCampaign(campaign.id), null, true)}
            disabled={pending}
            className="gap-1.5 text-muted-foreground hover:text-destructive"
          >
            {pending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Trash2 className="h-3 w-3" />}
            Delete
          </Button>
        )}
      </div>
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  );
}
