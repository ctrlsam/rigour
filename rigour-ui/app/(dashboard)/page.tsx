import { SearchHeader } from '../../components/SearchHeader';
import { FacetFilters } from '../../components/FacetFilters';
import { HostResults } from '../../components/HostResults';
import WorldMap from '../../components/ui/world-map';
import { searchHosts, getFacets, FacetCounts } from '../../lib/api';
import { Host } from '../../lib/types';

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function Home({ searchParams: searchParamsPromise }: PageProps) {
  const searchParams = await searchParamsPromise;
  
  let hosts: Host[] = [];
  let facets: FacetCounts = {};
  let error: string | null = null;

  try {
    // Build filter from search params
    const filter: Record<string, any> = {};

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

    if (selectedCountries.length > 0) {
      filter['asn.country'] = { $in: selectedCountries };
    }

    if (selectedASNs.length > 0) {
      const asnNumbers = selectedASNs.map(asn => parseInt(asn.replace('AS', '')));
      filter['asn.number'] = { $in: asnNumbers };
    }

    if (selectedServices.length > 0) {
      filter['services.protocol'] = { $in: selectedServices };
    }

    // Parse query syntax filters (e.g., from "services.protocol: ssh")
    if (searchParams.filter) {
      const filterParams = Array.isArray(searchParams.filter)
        ? searchParams.filter
        : [searchParams.filter];
      
      for (const filterParam of filterParams) {
        try {
          const parsedFilter = JSON.parse(filterParam);
          // Merge with existing filter
          Object.assign(filter, parsedFilter);
        } catch (e) {
          console.error('Failed to parse filter parameter:', e);
        }
      }
    }

    // Perform search
    const searchResult = await searchHosts(filter, 50);
    hosts = searchResult.hosts || [];

    // Fetch facets for the current filter to show accurate counts
    const facetsResult = await getFacets(filter);
    facets = facetsResult.facets || {};
  } catch (err) {
    console.error('Failed to fetch data:', err);
    error = err instanceof Error ? err.message : 'Failed to fetch data';
  }

  // Extract unique values for filters from facets
  const countries = Object.keys(facets.countries || {}).sort();
  const asns = Object.keys(facets.asns || {})
    .map(asnStr => {
      const parts = asnStr.split('-');
      return `AS${parts[0]}`;
    })
    .sort();
  const allServices = Object.keys(facets.services || {}).sort();

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
      <div className="dark min-h-screen bg-background">
        <div className="container mx-auto px-4 py-8">
          <div className="text-red-500">Error: {error}</div>
          <p className="text-sm text-gray-400 mt-2">
            Make sure the API is running at http://localhost:8080
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="dark min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <SearchHeader 
          initialQuery={
            typeof searchParams.query === 'string' ? searchParams.query : ''
          }
        />

        <div className="mt-8 grid grid-cols-1 lg:grid-cols-5 gap-6">
          <aside className="lg:col-span-1">
            <FacetFilters
              countries={countries}
              asns={asns}
              services={allServices}
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
              facets={facets}
            />
          </aside>

          <main className="lg:col-span-4 space-y-6">
            {hosts.length > 0 && <WorldMap dots={mapDots} />}
            <HostResults hosts={hosts} totalCount={hosts.length} />
          </main>
        </div>
      </div>
    </div>
  );
}