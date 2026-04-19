import { SearchHeader } from '../../components/SearchHeader';
import { FacetFilters } from '../../components/FacetFilters';
import { HostResults } from '../../components/HostResults';
import WorldMap from '../../components/ui/world-map';
import { searchHosts, getFacets, FacetCounts, API_BASE_URL } from '../../lib/api';
import { Host } from '../../lib/types';

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function Home({ searchParams: searchParamsPromise }: PageProps) {
  const searchParams = await searchParamsPromise;

  let hosts: Host[] = [];
  let facets: FacetCounts | undefined = undefined;
  let error: string | null = null;
  let nextPageToken: string | undefined = undefined;

  try {
    // Build filter from search params
    const filter: Record<string, unknown> = {};

    const pageToken = typeof searchParams.page_token === 'string' ? searchParams.page_token : undefined;

    // Parse filters from facet selections
    const selectedCountries = searchParams.countries
      ? Array.isArray(searchParams.countries)
        ? searchParams.countries
        : searchParams.countries.split(',')
      : [];
    const selectedASNs = searchParams.asns
      ? Array.isArray(searchParams.asns)
        ? searchParams.asns
        : searchParams.asns.split(',')
      : [];
    const selectedServices = searchParams.services
      ? Array.isArray(searchParams.services)
        ? searchParams.services
        : searchParams.services.split(',')
      : [];
    const selectedSeverities = searchParams.severity
      ? Array.isArray(searchParams.severity)
        ? searchParams.severity
        : searchParams.severity.split(',')
      : [];

    if (selectedCountries.length > 0) {
      filter['location.country_code'] = { $in: selectedCountries };
    }

    if (selectedASNs.length > 0) {
      const asnNumbers = selectedASNs.map(asn => parseInt(asn.replace('AS', '')));
      filter['asn.number'] = { $in: asnNumbers };
    }

    if (selectedServices.length > 0) {
      filter['services.protocol'] = { $in: selectedServices };
    }

    if (selectedSeverities.length === 1) {
      filter['vulns.max_severity'] = selectedSeverities[0];
    } else if (selectedSeverities.length > 1) {
      filter['vulns.max_severity'] = { $in: selectedSeverities };
    }

    // Parse query syntax filters (e.g., from "services.protocol: ssh")
    if (searchParams.filter) {
      const filterParams = Array.isArray(searchParams.filter)
        ? searchParams.filter
        : [searchParams.filter];

      for (const filterParam of filterParams) {
        try {
          const parsedFilter = JSON.parse(filterParam);
          Object.assign(filter, parsedFilter);
        } catch (e) {
          console.error('Failed to parse filter parameter:', e);
        }
      }
    }

    // Perform search — 10 hosts per page, cursor-based pagination
    const searchResult = await searchHosts(filter, 10, pageToken);
    hosts = searchResult.hosts || [];
    nextPageToken = searchResult.next_page_token;

    // Fetch facets for the current filter to show accurate counts
    const facetsResult = await getFacets(filter);
    facets = Object.keys(facetsResult.facets).length === 0
      ? { services: {}, countries: [], asns: [] }
      : facetsResult.facets;
  } catch (err) {
    console.error('Failed to fetch data:', err);
    error = err instanceof Error ? err.message : 'Failed to fetch data';
  }

  // World Map dots data
  const mapDots = hosts
    .filter(host => host.location.city !== 'Unknown')
    .map(host => ({
      start: {
        lat: host.location.coordinates[1],
        lng: host.location.coordinates[0],
        label: host.ip,
      },
      end: {
        lat: host.location.coordinates[1],
        lng: host.location.coordinates[0],
        label: host.ip,
      },
    }));

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-red-500">Error: {error}</div>
        <p className="text-sm text-gray-400 mt-2">
          Make sure the API is running at {API_BASE_URL}
        </p>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
        <SearchHeader
          initialQuery={
            typeof searchParams.query === 'string' ? searchParams.query : ''
          }
        />

        <div className="mt-8 grid grid-cols-1 lg:grid-cols-5 gap-6">
          <aside className="lg:col-span-1">
            {facets && (
            <FacetFilters
              facets={facets}
              selectedCountries={
                Array.isArray(searchParams.countries)
                  ? searchParams.countries
                  : searchParams.countries?.split(',') || []
              }
              selectedASNs={
                Array.isArray(searchParams.asns)
                  ? searchParams.asns
                  : searchParams.asns?.split(',') || []
              }
              selectedServices={
                Array.isArray(searchParams.services)
                  ? searchParams.services
                  : searchParams.services?.split(',') || []
              }
              selectedSeverities={
                Array.isArray(searchParams.severity)
                  ? searchParams.severity
                  : searchParams.severity?.split(',') || []
              }
            />
            )}
          </aside>

          <main className="lg:col-span-4 space-y-6">
            {hosts.length > 0 && <WorldMap dots={mapDots} />}
            <HostResults
              hosts={hosts}
              totalCount={hosts.length}
              currentPage={parseInt(searchParams.page as string || '1', 10)}
              nextPageToken={nextPageToken}
            />
          </main>
        </div>
    </div>
  );
}
