import { Host, Campaign, CreateCampaignRequest, Vulnerability, VulnerabilityTemplateSummary } from './types';

// Server-side uses BACKEND_URL directly; client-side uses relative paths (proxied by Next.js rewrites).
export const API_BASE_URL =
  typeof window === 'undefined'
    ? (process.env.BACKEND_URL || 'http://localhost:8080')
    : '';

export interface CountryFacet {
  code: string;
  name: string;
  count: number;
}

export interface ASNFacet {
  code: number;
  name: string;
  count: number;
}

export interface SearchResponse {
  hosts: Host[];
  facets?: FacetCounts;
  next_page_token?: string;
}

export interface FacetCounts {
  services: Record<string, number>;
  countries: CountryFacet[];
  asns: ASNFacet[];
}

export interface FacetsResponse {
  facets: FacetCounts;
}

/**
 * Search for hosts with optional filters and pagination
 */
export async function searchHosts(
  filter?: Record<string, any>,
  limit: number = 20,
  pageToken?: string
): Promise<SearchResponse> {
  const params = new URLSearchParams();

  if (filter && Object.keys(filter).length > 0) {
    params.append('filter', JSON.stringify(filter));
  }

  if (limit) {
    params.append('limit', limit.toString());
  }

  if (pageToken) {
    params.append('page_token', pageToken);
  }

  const url = `${API_BASE_URL}/api/hosts/search?${params.toString()}`;

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return response.json();
}

/**
 * Get facet aggregations with optional filters
 */
export async function getFacets(
  filter?: Record<string, any>
): Promise<FacetsResponse> {
  const params = new URLSearchParams();

  if (filter && Object.keys(filter).length > 0) {
    params.append('filter', JSON.stringify(filter));
  }

  const url = `${API_BASE_URL}/api/facets?${params.toString()}`;

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return response.json();
}

export async function listCampaigns(): Promise<Campaign[]> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns`, {
    cache: 'no-store',
  });
  if (!response.ok) throw new Error(`API error: ${response.status}`);
  return response.json();
}

export async function getCampaign(id: string): Promise<Campaign> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns/${id}`, {
    cache: 'no-store',
  });
  if (!response.ok) throw new Error(`API error: ${response.status}`);
  return response.json();
}

export async function createCampaign(req: CreateCampaignRequest): Promise<Campaign> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${response.status}`);
  }
  return response.json();
}

export async function startCampaign(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns/${id}/start`, { method: 'POST' });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${response.status}`);
  }
}

export async function pauseCampaign(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns/${id}/pause`, { method: 'POST' });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${response.status}`);
  }
}

export async function resumeCampaign(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns/${id}/resume`, { method: 'POST' });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${response.status}`);
  }
}

export async function cancelCampaign(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns/${id}/cancel`, { method: 'POST' });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${response.status}`);
  }
}

export async function deleteCampaign(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/admin/campaigns/${id}`, { method: 'DELETE' });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${response.status}`);
  }
}

/**
 * Get vulnerabilities grouped by template, sorted by severity (critical first).
 */
export async function getVulnerabilitySummary(severity?: string): Promise<VulnerabilityTemplateSummary[]> {
  const params = new URLSearchParams();
  if (severity) params.append('severity', severity);
  const url = `${API_BASE_URL}/api/vulnerabilities/summary?${params.toString()}`;
  const response = await fetch(url, { method: 'GET', headers: { 'Content-Type': 'application/json' } });
  if (!response.ok) throw new Error(`API error: ${response.status}`);
  return response.json();
}

/**
 * Get all vulnerabilities for a host by IP. Returns empty array on 404 (vuln scan disabled).
 */
export async function getVulnsByIP(ip: string): Promise<Vulnerability[]> {
  const url = `${API_BASE_URL}/api/hosts/${ip}/vulnerabilities`;
  const response = await fetch(url, { method: 'GET', headers: { 'Content-Type': 'application/json' } });
  if (response.status === 404) return [];
  if (!response.ok) throw new Error(`API error: ${response.status}`);
  return response.json();
}

/**
 * Get a single host by IP address
 */
export async function getHostByIP(ip: string): Promise<Host> {
  const url = `${API_BASE_URL}/api/hosts/${ip}`;

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return response.json();
}
