export type CampaignStatus =
  | 'pending'
  | 'running'
  | 'paused'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface CampaignProgress {
  total_cidrs: number;
  completed_cidrs: number;
  current_cidr?: string;
  hosts_discovered: number;
  services_found: number;
}

export interface Campaign {
  id: string;
  name: string;
  description?: string;
  status: CampaignStatus;
  cidrs: string[] | null;
  country_code?: string;
  ports?: string;
  top_ports?: string;
  service_filter?: string[];
  scan_type: string;
  rate: number;
  retries: number;
  timeout: number;
  fast_mode: boolean;
  use_udp: boolean;
  run_vuln_scan: boolean;
  workers?: number;
  queue_size?: number;
  concurrent_cidrs?: number;
  progress: CampaignProgress;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  updated_at: string;
  error?: string;
}

export interface CreateCampaignRequest {
  name: string;
  description?: string;
  cidrs?: string[];
  country_code?: string;
  ports?: string;
  top_ports?: string;
  service_filter?: string[];
  scan_type?: string;
  rate?: number;
  retries?: number;
  timeout?: number;
  fast_mode?: boolean;
  use_udp?: boolean;
  run_vuln_scan?: boolean;
  vuln_scan_templates?: string[];
}

export type SeverityLevel = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface VulnSummary {
  count: number;
  max_severity?: SeverityLevel;
  cve_ids?: string[];
}

export interface VulnerabilityTemplateSummary {
  template_id: string;
  name: string;
  severity: SeverityLevel;
  cvss?: number;
  cve_ids?: string[];
  description?: string;
  affected_host_count: number;
}

export interface Vulnerability {
  id: string;
  ip: string;
  port: number;
  protocol: string;
  template_id: string;
  name: string;
  severity: SeverityLevel;
  description?: string;
  cvss?: number;
  cve_ids?: string[];
  matched_at: string;
  discovered_at: string;
  campaign_id?: string;
}

export interface Service {
  ip: string;
  port: number;
  protocol: string;
  tls: boolean;
  transport: string;
  last_scan: string;
  https?: {
    status: string;
    statusCode: number;
    responseHeaders: Record<string, string[]>;
  };
  http?: {
    status: string;
    statusCode: number;
    responseHeaders: Record<string, string[]>;
  };
  ssh?: {
    banner: string;
  };
}

export interface Host {
  id: string;
  ip: string;
  ip_int: number;
  asn: {
    number: number;
    organization: string;
    country: string;
  };
  location: {
    coordinates: [number, number];
    city: string;
    timezone: string;
    country_code: string;
    country_name: string;
  };
  first_seen: string;
  last_seen: string;
  services: Service[];
  vulns?: VulnSummary;
}

