'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import {
  Play,
  Pause,
  Square,
  Trash2,
  Plus,
  RotateCcw,
  ChevronRight,
  Loader2,
  Target,
} from 'lucide-react';
import { Campaign, CreateCampaignRequest, CampaignStatus } from '../lib/types';
import { CAMPAIGN_STATUS_STYLES, CAMPAIGN_STATUS_LABELS } from '../lib/constants';
import * as api from '../lib/api';
import { Card, CardContent } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { cn } from '../lib/utils';

const EMPTY_FORM = {
  name: '',
  description: '',
  targetType: 'cidrs' as 'cidrs' | 'country',
  cidrs: '',
  country_code: '',
  portType: 'top' as 'top' | 'custom',
  top_ports: '1000',
  ports: '',
  scan_type: 'c',
  rate: '',
  retries: '',
  timeout: '',
  fast_mode: false,
  use_udp: false,
  service_filter: '',
  run_vuln_scan: false,
  vuln_scan_templates: '',
};

export function CampaignList() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ ...EMPTY_FORM });
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState('');
  const [actionError, setActionError] = useState<Record<string, string>>({});

  const fetchCampaigns = useCallback(async () => {
    try {
      const data = await api.listCampaigns();
      setCampaigns(data ?? []);
    } catch {
      // keep stale data on error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCampaigns();
  }, [fetchCampaigns]);

  // Poll while any campaign is active
  useEffect(() => {
    const hasActive = campaigns.some(
      (c) => c.status === 'running' || c.status === 'paused'
    );
    if (!hasActive) return;
    const id = setInterval(fetchCampaigns, 5000);
    return () => clearInterval(id);
  }, [campaigns, fetchCampaigns]);

  async function handleCreate() {
    setFormError('');
    if (!form.name.trim()) {
      setFormError('Name is required');
      return;
    }
    if (form.targetType === 'cidrs' && !form.cidrs.trim()) {
      setFormError('At least one CIDR is required');
      return;
    }
    if (form.targetType === 'country' && !form.country_code.trim()) {
      setFormError('Country code is required');
      return;
    }

    const req: CreateCampaignRequest = { name: form.name.trim() };
    if (form.description.trim()) req.description = form.description.trim();

    if (form.targetType === 'cidrs') {
      req.cidrs = form.cidrs
        .split(/[\n,]/)
        .map((s) => s.trim())
        .filter(Boolean);
    } else {
      req.country_code = form.country_code.trim().toUpperCase();
    }

    if (form.portType === 'top') {
      req.top_ports = form.top_ports;
    } else if (form.ports.trim()) {
      req.ports = form.ports.trim();
    }

    if (form.scan_type) req.scan_type = form.scan_type;
    if (form.rate) req.rate = parseInt(form.rate);
    if (form.retries) req.retries = parseInt(form.retries);
    if (form.timeout) req.timeout = parseInt(form.timeout);
    if (form.fast_mode) req.fast_mode = true;
    if (form.use_udp) req.use_udp = true;
    if (form.service_filter.trim()) {
      req.service_filter = form.service_filter.split(',').map((s) => s.trim()).filter(Boolean);
    }
    if (form.run_vuln_scan) {
      req.run_vuln_scan = true;
      const tags = (form.vuln_scan_templates ?? '').trim();
      if (tags) {
        req.vuln_scan_templates = tags.split(',').map((s) => s.trim()).filter(Boolean);
      }
    }

    setSubmitting(true);
    try {
      await api.createCampaign(req);
      setShowCreate(false);
      setForm({ ...EMPTY_FORM });
      await fetchCampaigns();
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'Failed to create campaign');
    } finally {
      setSubmitting(false);
    }
  }

  async function handleAction(id: string, action: () => Promise<void>, optimisticStatus?: CampaignStatus) {
    setActionError((prev) => ({ ...prev, [id]: '' }));
    if (optimisticStatus) {
      setCampaigns((prev) =>
        prev.map((c) => (c.id === id ? { ...c, status: optimisticStatus } : c))
      );
    }
    try {
      await action();
      await fetchCampaigns();
    } catch (e) {
      // Roll back optimistic update on failure
      await fetchCampaigns();
      setActionError((prev) => ({
        ...prev,
        [id]: e instanceof Error ? e.message : 'Action failed',
      }));
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Campaigns</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Targeted scan jobs across CIDR ranges
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          New Campaign
        </Button>
      </div>

      {campaigns.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <Target className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground mb-2">No campaigns yet</p>
          <p className="text-sm text-muted-foreground mb-6">
            Create a campaign to start a targeted network scan
          </p>
          <Button onClick={() => setShowCreate(true)} variant="outline" className="gap-2">
            <Plus className="h-4 w-4" />
            Create your first campaign
          </Button>
        </div>
      ) : (
        <div className="grid gap-4">
          {campaigns.map((campaign) => (
            <CampaignCard
              key={campaign.id}
              campaign={campaign}
              error={actionError[campaign.id]}
              onStart={() => handleAction(campaign.id, () => api.startCampaign(campaign.id), 'running')}
              onPause={() => handleAction(campaign.id, () => api.pauseCampaign(campaign.id), 'paused')}
              onResume={() => handleAction(campaign.id, () => api.resumeCampaign(campaign.id), 'running')}
              onCancel={() => handleAction(campaign.id, () => api.cancelCampaign(campaign.id), 'cancelled')}
              onDelete={() => handleAction(campaign.id, () => api.deleteCampaign(campaign.id))}
            />
          ))}
        </div>
      )}

      {showCreate && (
        <CreateCampaignModal
          form={form}
          onChange={setForm}
          onSubmit={handleCreate}
          onClose={() => {
            setShowCreate(false);
            setForm({ ...EMPTY_FORM });
            setFormError('');
          }}
          submitting={submitting}
          error={formError}
        />
      )}
    </div>
  );
}

interface CampaignCardProps {
  campaign: Campaign;
  error?: string;
  onStart: () => void;
  onPause: () => void;
  onResume: () => void;
  onCancel: () => void;
  onDelete: () => void;
}

function CampaignCard({
  campaign,
  error,
  onStart,
  onPause,
  onResume,
  onCancel,
  onDelete,
}: CampaignCardProps) {
  const { progress } = campaign;
  const pct =
    progress.total_cidrs > 0
      ? Math.round((progress.completed_cidrs / progress.total_cidrs) * 100)
      : 0;

  return (
    <Card className="bg-card border-border">
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-2 flex-wrap">
              <Link
                href={`/campaigns/${campaign.id}`}
                className="text-base font-semibold hover:text-blue-400 transition-colors truncate"
              >
                {campaign.name}
              </Link>
              <Badge
                variant="outline"
                className={cn('text-xs shrink-0', CAMPAIGN_STATUS_STYLES[campaign.status])}
              >
                {campaign.status === 'running' && (
                  <span className="mr-1.5 h-1.5 w-1.5 rounded-full bg-blue-400 animate-pulse inline-block" />
                )}
                {CAMPAIGN_STATUS_LABELS[campaign.status]}
              </Badge>
            </div>

            {campaign.description && (
              <p className="text-sm text-muted-foreground mb-3 truncate">{campaign.description}</p>
            )}

            {(campaign.status === 'running' || campaign.status === 'paused') && (
              <div className="mb-3 space-y-1.5">
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>
                    {progress.completed_cidrs} / {progress.total_cidrs} CIDRs
                  </span>
                  <span>{pct}%</span>
                </div>
                <div className="w-full h-1.5 bg-secondary rounded-full overflow-hidden">
                  <div
                    className={cn(
                      'h-full rounded-full transition-all duration-300',
                      campaign.status === 'running' ? 'bg-blue-500' : 'bg-yellow-500'
                    )}
                    style={{ width: `${pct}%` }}
                  />
                </div>
                {progress.current_cidr && (
                  <p className="text-xs text-muted-foreground font-mono truncate">
                    Scanning {progress.current_cidr}
                  </p>
                )}
              </div>
            )}

            <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
              <span>{campaign.cidrs?.length ? `${campaign.cidrs.length} CIDRs` : campaign.country_code}</span>
              <span>{progress.services_found.toLocaleString()} services</span>
              <span>
                Created {new Date(campaign.created_at).toLocaleDateString()}
              </span>
              {campaign.ports && <span className="font-mono">ports: {campaign.ports}</span>}
              {campaign.top_ports && <span className="font-mono">top {campaign.top_ports}</span>}
              {campaign.run_vuln_scan && (
                <span className="text-purple-400 font-medium">vuln scan</span>
              )}
            </div>

            {campaign.error && (
              <p className="mt-2 text-xs text-red-400 font-mono">{campaign.error}</p>
            )}
            {error && (
              <p className="mt-2 text-xs text-red-400">{error}</p>
            )}
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <Link href={`/campaigns/${campaign.id}`}>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                <ChevronRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
        </div>

        <div className="flex gap-2 mt-4 pt-4 border-t border-border flex-wrap">
          {campaign.status === 'pending' && (
            <Button size="sm" onClick={onStart} className="gap-1.5 h-8">
              <Play className="h-3 w-3" />
              Start
            </Button>
          )}
          {campaign.status === 'running' && (
            <Button size="sm" variant="outline" onClick={onPause} className="gap-1.5 h-8">
              <Pause className="h-3 w-3" />
              Pause
            </Button>
          )}
          {campaign.status === 'paused' && (
            <Button size="sm" onClick={onResume} className="gap-1.5 h-8">
              <RotateCcw className="h-3 w-3" />
              Resume
            </Button>
          )}
          {(campaign.status === 'running' || campaign.status === 'paused') && (
            <Button size="sm" variant="outline" onClick={onCancel} className="gap-1.5 h-8 text-destructive border-destructive/30 hover:bg-destructive/10">
              <Square className="h-3 w-3" />
              Cancel
            </Button>
          )}
          {(campaign.status === 'completed' || campaign.status === 'failed' || campaign.status === 'cancelled' || campaign.status === 'pending') && (
            <Button
              size="sm"
              variant="ghost"
              onClick={onDelete}
              className="gap-1.5 h-8 text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-3 w-3" />
              Delete
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

type FormState = typeof EMPTY_FORM;

interface CreateCampaignModalProps {
  form: FormState;
  onChange: (form: FormState) => void;
  onSubmit: () => void;
  onClose: () => void;
  submitting: boolean;
  error: string;
}

function CreateCampaignModal({
  form,
  onChange,
  onSubmit,
  onClose,
  submitting,
  error,
}: CreateCampaignModalProps) {
  const set = (patch: Partial<FormState>) => onChange({ ...form, ...patch });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative z-10 w-full max-w-lg bg-card border border-border rounded-xl shadow-2xl max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between p-6 border-b border-border">
          <h2 className="text-lg font-semibold">New Campaign</h2>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground transition-colors text-xl leading-none"
          >
            ×
          </button>
        </div>

        <div className="overflow-y-auto p-6 space-y-5 flex-1">
          {/* Name */}
          <div className="space-y-1.5">
            <Label htmlFor="c-name">Name *</Label>
            <Input
              id="c-name"
              placeholder="My Scan Campaign"
              value={form.name}
              onChange={(e) => set({ name: e.target.value })}
            />
          </div>

          {/* Description */}
          <div className="space-y-1.5">
            <Label htmlFor="c-desc">Description</Label>
            <Input
              id="c-desc"
              placeholder="Optional description"
              value={form.description}
              onChange={(e) => set({ description: e.target.value })}
            />
          </div>

          {/* Target type toggle */}
          <div className="space-y-2">
            <Label>Target</Label>
            <div className="flex rounded-lg border border-border overflow-hidden text-sm">
              <button
                type="button"
                onClick={() => set({ targetType: 'cidrs' })}
                className={cn(
                  'flex-1 py-2 transition-colors',
                  form.targetType === 'cidrs'
                    ? 'bg-secondary text-foreground'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                CIDR Ranges
              </button>
              <button
                type="button"
                onClick={() => set({ targetType: 'country' })}
                className={cn(
                  'flex-1 py-2 transition-colors border-l border-border',
                  form.targetType === 'country'
                    ? 'bg-secondary text-foreground'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                Country Code
              </button>
            </div>

            {form.targetType === 'cidrs' ? (
              <textarea
                className="w-full min-h-[80px] rounded-md border border-border bg-secondary px-3 py-2 text-sm font-mono placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring resize-y"
                placeholder={'192.168.1.0/24\n10.0.0.0/8'}
                value={form.cidrs}
                onChange={(e) => set({ cidrs: e.target.value })}
              />
            ) : (
              <Input
                placeholder="AU"
                maxLength={2}
                value={form.country_code}
                onChange={(e) => set({ country_code: e.target.value })}
                className="font-mono uppercase"
              />
            )}
          </div>

          {/* Port config */}
          <div className="space-y-2">
            <Label>Ports</Label>
            <div className="flex rounded-lg border border-border overflow-hidden text-sm">
              <button
                type="button"
                onClick={() => set({ portType: 'top' })}
                className={cn(
                  'flex-1 py-2 transition-colors',
                  form.portType === 'top'
                    ? 'bg-secondary text-foreground'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                Top Ports
              </button>
              <button
                type="button"
                onClick={() => set({ portType: 'custom' })}
                className={cn(
                  'flex-1 py-2 transition-colors border-l border-border',
                  form.portType === 'custom'
                    ? 'bg-secondary text-foreground'
                    : 'text-muted-foreground hover:text-foreground'
                )}
              >
                Custom
              </button>
            </div>

            {form.portType === 'top' ? (
              <select
                className="w-full rounded-md border border-border bg-secondary px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                value={form.top_ports}
                onChange={(e) => set({ top_ports: e.target.value })}
              >
                <option value="100">Top 100</option>
                <option value="1000">Top 1000</option>
                <option value="full">Full</option>
              </select>
            ) : (
              <Input
                placeholder='80,443,8080-8090'
                value={form.ports}
                onChange={(e) => set({ ports: e.target.value })}
                className="font-mono"
              />
            )}
          </div>

          {/* Advanced settings */}
          <details className="group">
            <summary className="text-sm text-muted-foreground cursor-pointer hover:text-foreground transition-colors list-none flex items-center gap-1.5">
              <ChevronRight className="h-3.5 w-3.5 transition-transform group-open:rotate-90" />
              Advanced settings
            </summary>
            <div className="mt-4 space-y-4 pl-5">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label>Scan type</Label>
                  <select
                    className="w-full rounded-md border border-border bg-secondary px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    value={form.scan_type}
                    onChange={(e) => set({ scan_type: e.target.value })}
                  >
                    <option value="c">Connect (c)</option>
                    <option value="s">SYN (s)</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="c-rate">Rate (pps)</Label>
                  <Input
                    id="c-rate"
                    type="number"
                    placeholder="500"
                    value={form.rate}
                    onChange={(e) => set({ rate: e.target.value })}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="c-retries">Retries</Label>
                  <Input
                    id="c-retries"
                    type="number"
                    placeholder="1"
                    value={form.retries}
                    onChange={(e) => set({ retries: e.target.value })}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="c-timeout">Timeout (ms)</Label>
                  <Input
                    id="c-timeout"
                    type="number"
                    placeholder="1000"
                    value={form.timeout}
                    onChange={(e) => set({ timeout: e.target.value })}
                  />
                </div>
              </div>
              <div className="flex gap-6">
                <label className="flex items-center gap-2 text-sm cursor-pointer" title="Only probe protocols whose default port matches the discovered port. Skips exhaustive protocol guessing — useful when scanning a well-known port.">
                  <input
                    type="checkbox"
                    checked={form.fast_mode}
                    onChange={(e) => set({ fast_mode: e.target.checked })}
                    className="rounded border-border"
                  />
                  Port-match only
                </label>
                <label className="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="checkbox"
                    checked={form.use_udp}
                    onChange={(e) => set({ use_udp: e.target.checked })}
                    className="rounded border-border"
                  />
                  UDP plugins
                </label>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="c-filter">Protocol filter</Label>
                <Input
                  id="c-filter"
                  placeholder="http, ssh, minecraft-java (empty = all)"
                  value={form.service_filter}
                  onChange={(e) => set({ service_filter: e.target.value })}
                  className="font-mono text-xs"
                />
                <p className="text-xs text-muted-foreground">Comma-separated protocol names. Limits fingerprinting to these protocols only.</p>
              </div>

              <div className="space-y-3 pt-1 border-t border-border">
                <label className="flex items-center gap-2 text-sm cursor-pointer" title="Run Nuclei vulnerability scanning on each discovered service. Requires the server to be started with --vuln-scan.">
                  <input
                    type="checkbox"
                    checked={form.run_vuln_scan}
                    onChange={(e) => set({ run_vuln_scan: e.target.checked })}
                    className="rounded border-border"
                  />
                  <span>Vulnerability scanning <span className="text-muted-foreground">(requires <code className="text-xs">--vuln-scan</code> server flag)</span></span>
                </label>
                {form.run_vuln_scan && (
                  <div className="space-y-1.5 pl-6">
                    <Label htmlFor="c-vuln-tags">Nuclei template tags</Label>
                    <Input
                      id="c-vuln-tags"
                      placeholder="cve, rce, default-login, exposure"
                      value={form.vuln_scan_templates ?? ''}
                      onChange={(e) => set({ vuln_scan_templates: e.target.value })}
                      className="font-mono text-xs"
                    />
                    <p className="text-xs text-muted-foreground">Comma-separated tags. Leave empty to use the server default.</p>
                  </div>
                )}
              </div>
            </div>
          </details>

          {error && <p className="text-sm text-red-400">{error}</p>}
        </div>

        <div className="flex justify-end gap-3 p-6 border-t border-border">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={submitting} className="gap-2">
            {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
            Create Campaign
          </Button>
        </div>
      </div>
    </div>
  );
}
