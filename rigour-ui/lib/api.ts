import { Host } from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

export interface SearchResponse {
  hosts: Host[];
  facets?: FacetCounts;
  next_page_token?: string;
}

export interface FacetCounts {
  services?: Record<string, number>;
  countries?: Record<string, number>;
  asns?: Record<string, number>;
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
