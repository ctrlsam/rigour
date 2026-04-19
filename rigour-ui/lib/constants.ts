import { CampaignStatus, SeverityLevel } from './types';

export const CAMPAIGN_STATUS_STYLES: Record<CampaignStatus, string> = {
  pending: 'bg-secondary text-secondary-foreground border-border',
  running: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  paused: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  completed: 'bg-green-500/20 text-green-400 border-green-500/30',
  failed: 'bg-red-500/20 text-red-400 border-red-500/30',
  cancelled: 'bg-secondary text-muted-foreground border-border',
};

export const CAMPAIGN_STATUS_LABELS: Record<CampaignStatus, string> = {
  pending: 'Pending',
  running: 'Running',
  paused: 'Paused',
  completed: 'Completed',
  failed: 'Failed',
  cancelled: 'Cancelled',
};

export const SEVERITY_CLASSES: Record<SeverityLevel, string> = {
  critical: 'bg-red-500/20 text-red-400 border-red-500/30',
  high:     'bg-orange-500/20 text-orange-400 border-orange-500/30',
  medium:   'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  low:      'bg-blue-500/20 text-blue-400 border-blue-500/30',
  info:     'bg-gray-500/20 text-gray-400 border-gray-500/30',
};
