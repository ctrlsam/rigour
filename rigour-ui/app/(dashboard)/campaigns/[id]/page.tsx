import Link from 'next/link';
import { getCampaign } from '../../../../lib/api';
import { Campaign, CampaignStatus } from '../../../../lib/types';
import { Card, CardContent, CardHeader } from '../../../../components/ui/card';
import { Badge } from '../../../../components/ui/badge';
import { Button } from '../../../../components/ui/button';
import { CampaignDetailActions } from '../../../../components/CampaignDetailActions';
import { AlertCircle, ChevronLeft, ShieldCheck, ShieldOff, Target } from 'lucide-react';
import { cn } from '../../../../lib/utils';

const STATUS_STYLES: Record<CampaignStatus, string> = {
  pending: 'bg-secondary text-secondary-foreground border-border',
  running: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  paused: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  completed: 'bg-green-500/20 text-green-400 border-green-500/30',
  failed: 'bg-red-500/20 text-red-400 border-red-500/30',
  cancelled: 'bg-secondary text-muted-foreground border-border',
};

const STATUS_LABELS: Record<CampaignStatus, string> = {
  pending: 'Pending',
  running: 'Running',
  paused: 'Paused',
  completed: 'Completed',
  failed: 'Failed',
  cancelled: 'Cancelled',
};

export default async function CampaignDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  let campaign: Campaign | null = null;
  let error: string | null = null;

  try {
    campaign = await getCampaign(id);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to fetch campaign';
  }

  if (error || !campaign) {
    return (
      <div className="p-6">
        <div className="max-w-4xl mx-auto space-y-6">
          <Link href="/campaigns">
            <Button size="sm" variant="outline" className="gap-2">
              <ChevronLeft className="h-4 w-4" />
              Back to Campaigns
            </Button>
          </Link>
          <div className="text-center py-12">
            <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
            <p className="text-muted-foreground">{error || 'Campaign not found'}</p>
          </div>
        </div>
      </div>
    );
  }

  const { progress } = campaign;
  const pct =
    progress.total_cidrs > 0
      ? Math.round((progress.completed_cidrs / progress.total_cidrs) * 100)
      : 0;

  return (
    <div className="p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <Link href="/campaigns">
            <Button size="sm" variant="outline" className="gap-2">
              <ChevronLeft className="h-4 w-4" />
              Back to Campaigns
            </Button>
          </Link>
          <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">{campaign.id}</code>
        </div>

        {/* Summary */}
        <Card className="bg-card border-border">
          <CardHeader className="pb-4">
            <div className="flex items-start justify-between gap-4 flex-wrap">
              <div className="space-y-2">
                <div className="flex items-center gap-3 flex-wrap">
                  <h1 className="text-2xl font-bold">{campaign.name}</h1>
                  <Badge
                    variant="outline"
                    className={cn('text-sm', STATUS_STYLES[campaign.status])}
                  >
                    {campaign.status === 'running' && (
                      <span className="mr-1.5 h-1.5 w-1.5 rounded-full bg-blue-400 animate-pulse inline-block" />
                    )}
                    {STATUS_LABELS[campaign.status]}
                  </Badge>
                </div>
                {campaign.description && (
                  <p className="text-muted-foreground">{campaign.description}</p>
                )}
              </div>
              <CampaignDetailActions campaign={campaign} />
            </div>
          </CardHeader>
        </Card>

        {/* Progress */}
        {(campaign.status === 'running' ||
          campaign.status === 'paused' ||
          campaign.status === 'completed') && (
          <Card className="bg-card border-border">
            <CardHeader>
              <h2 className="text-base font-semibold flex items-center gap-2">
                <Target className="h-4 w-4" />
                Progress
              </h2>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-1.5">
                <div className="flex justify-between text-sm text-muted-foreground">
                  <span>{progress.completed_cidrs} / {progress.total_cidrs} CIDRs</span>
                  <span>{pct}%</span>
                </div>
                <div className="w-full h-2 bg-secondary rounded-full overflow-hidden">
                  <div
                    className={cn(
                      'h-full rounded-full transition-all duration-300',
                      campaign.status === 'running'
                        ? 'bg-blue-500'
                        : campaign.status === 'paused'
                        ? 'bg-yellow-500'
                        : 'bg-green-500'
                    )}
                    style={{ width: `${pct}%` }}
                  />
                </div>
                {progress.current_cidr && (
                  <p className="text-xs text-muted-foreground font-mono">
                    Currently scanning: {progress.current_cidr}
                  </p>
                )}
              </div>

              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-2 border-t border-border">
                <Stat label="Total CIDRs" value={progress.total_cidrs} />
                <Stat label="Completed" value={progress.completed_cidrs} />
                <Stat label="Services Found" value={progress.services_found.toLocaleString()} />
                <Stat label="Hosts Discovered" value={progress.hosts_discovered.toLocaleString()} />
              </div>
            </CardContent>
          </Card>
        )}

        {/* Error */}
        {campaign.error && (
          <Card className="bg-card border-red-500/30">
            <CardContent className="p-4">
              <p className="text-sm text-red-400 font-mono">{campaign.error}</p>
            </CardContent>
          </Card>
        )}

        {/* Configuration */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h2 className="text-base font-semibold">Configuration</h2>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              <ConfigItem label="Scan type" value={campaign.scan_type === 'c' ? 'Connect' : 'SYN'} />
              <ConfigItem label="Ports" value={campaign.ports || `Top ${campaign.top_ports ?? '1000'}`} />
              <ConfigItem label="Rate" value={campaign.rate ? `${campaign.rate.toLocaleString()} pps` : 'Default (500 pps)'} />
              <ConfigItem label="Retries" value={campaign.retries || 'Default'} />
              <ConfigItem label="Timeout" value={campaign.timeout ? `${campaign.timeout}ms` : 'Default (1000ms)'} />
              <ConfigItem label="Fast mode" value={campaign.fast_mode ? 'Yes' : 'No'} />
              <ConfigItem label="UDP" value={campaign.use_udp ? 'Enabled' : 'Disabled'} />
              {campaign.country_code && <ConfigItem label="Country" value={campaign.country_code} />}
              {campaign.workers ? <ConfigItem label="Workers" value={campaign.workers} /> : null}
              {campaign.queue_size ? <ConfigItem label="Queue size" value={campaign.queue_size} /> : null}
              {campaign.concurrent_cidrs ? <ConfigItem label="Concurrent CIDRs" value={campaign.concurrent_cidrs} /> : null}
            </div>

            {campaign.service_filter && campaign.service_filter.length > 0 && (
              <div className="pt-3 border-t border-border space-y-1">
                <p className="text-xs text-muted-foreground uppercase tracking-wider">Protocol filter</p>
                <div className="flex flex-wrap gap-1.5">
                  {campaign.service_filter.map((p) => (
                    <span key={p} className="text-xs font-mono bg-secondary px-2 py-0.5 rounded border border-border">
                      {p}
                    </span>
                  ))}
                </div>
              </div>
            )}

            <div className="pt-3 border-t border-border">
              <div className={cn(
                'inline-flex items-center gap-2 text-sm px-3 py-1.5 rounded-md border',
                campaign.run_vuln_scan
                  ? 'bg-purple-500/10 text-purple-400 border-purple-500/30'
                  : 'bg-secondary text-muted-foreground border-border'
              )}>
                {campaign.run_vuln_scan
                  ? <><ShieldCheck className="h-4 w-4" /> Vulnerability scanning enabled</>
                  : <><ShieldOff className="h-4 w-4" /> Vulnerability scanning disabled</>}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Timeline */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h2 className="text-base font-semibold">Timeline</h2>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Stat label="Created" value={new Date(campaign.created_at).toLocaleString()} small />
              {campaign.started_at && (
                <Stat label="Started" value={new Date(campaign.started_at).toLocaleString()} small />
              )}
              {campaign.completed_at && (
                <Stat label="Completed" value={new Date(campaign.completed_at).toLocaleString()} small />
              )}
            </div>
          </CardContent>
        </Card>

        {/* CIDRs */}
        {campaign.cidrs && campaign.cidrs.length > 0 && (
          <Card className="bg-card border-border">
            <CardHeader>
              <h2 className="text-base font-semibold">
                CIDR Ranges ({campaign.cidrs?.length ?? 0})
              </h2>
            </CardHeader>
            <CardContent>
              <div className="max-h-64 overflow-y-auto space-y-1">
                {campaign.cidrs.map((cidr) => (
                  <div key={cidr} className="flex items-center gap-2">
                    <span className="text-xs font-mono text-muted-foreground">
                      {cidr === progress.current_cidr ? (
                        <span className="text-blue-400">{cidr} ←</span>
                      ) : (
                        cidr
                      )}
                    </span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}

function Stat({ label, value, small }: { label: string; value: string | number; small?: boolean }) {
  return (
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground uppercase tracking-wider">{label}</p>
      <p className={cn('font-medium', small ? 'text-sm' : 'text-lg')}>{value}</p>
    </div>
  );
}

function ConfigItem({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground uppercase tracking-wider">{label}</p>
      <p className="text-sm font-mono">{value}</p>
    </div>
  );
}
