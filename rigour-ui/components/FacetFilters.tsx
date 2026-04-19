'use client';

import { useState, useTransition } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Checkbox } from './ui/checkbox';
import { Label } from './ui/label';
import { ScrollArea } from './ui/scroll-area';
import { Button } from './ui/button';
import { ChevronDown, ChevronRight, Loader2 } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import { FacetCounts } from '../lib/api';

const SEVERITY_OPTIONS = ['critical', 'high', 'medium', 'low', 'info'] as const;

interface FacetFiltersProps {
  facets: FacetCounts;
  selectedCountries: string[];
  selectedASNs: string[];
  selectedServices: string[];
  selectedSeverities?: string[];
}

export function FacetFilters({
  facets,
  selectedCountries: initialCountries,
  selectedASNs: initialASNs,
  selectedServices: initialServices,
  selectedSeverities: initialSeverities = [],
}: FacetFiltersProps) {
  const [expandedSections, setExpandedSections] = useState({
    severity: true,
    countries: true,
    asns: true,
    services: true,
  });
  // Local state for selections before applying
  const [tempSeverities, setTempSeverities] = useState(initialSeverities);
  const [tempCountries, setTempCountries] = useState(initialCountries);
  const [tempASNs, setTempASNs] = useState(initialASNs);
  const [tempServices, setTempServices] = useState(initialServices);

  const [isPending, startTransition] = useTransition();
  const router = useRouter();
  const searchParams = useSearchParams();

  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections(prev => ({
      ...prev,
      [section]: !prev[section],
    }));
  };

  const applyFilters = () => {
    startTransition(() => {
      const params = new URLSearchParams(searchParams.toString());

      // Reset pagination when filters change
      params.delete('page');
      params.delete('page_token');

      // Overwrite facet params with current selections
      params.delete('severity');
      params.delete('countries');
      params.delete('asns');
      params.delete('services');

      if (tempSeverities.length > 0) params.set('severity', tempSeverities.join(','));
      if (tempCountries.length > 0) params.set('countries', tempCountries.join(','));
      if (tempASNs.length > 0) params.set('asns', tempASNs.join(','));
      if (tempServices.length > 0) params.set('services', tempServices.join(','));

      router.push(`?${params.toString()}`);
    });
  };

  const handleSeverityToggle = (sev: string) => {
    setTempSeverities(prev =>
      prev.includes(sev)
        ? prev.filter(s => s !== sev)
        : [...prev, sev]
    );
  };

  const handleCountryToggle = (country: string) => {
    setTempCountries(prev =>
      prev.includes(country)
        ? prev.filter(c => c !== country)
        : [...prev, country]
    );
  };

  const handleASNToggle = (asn: string) => {
    setTempASNs(prev =>
      prev.includes(asn)
        ? prev.filter(a => a !== asn)
        : [...prev, asn]
    );
  };

  const handleServiceToggle = (service: string) => {
    setTempServices(prev =>
      prev.includes(service)
        ? prev.filter(s => s !== service)
        : [...prev, service]
    );
  };

  // Check if there are unsaved changes
  const hasChanges =
    JSON.stringify(tempSeverities) !== JSON.stringify(initialSeverities) ||
    JSON.stringify(tempCountries) !== JSON.stringify(initialCountries) ||
    JSON.stringify(tempASNs) !== JSON.stringify(initialASNs) ||
    JSON.stringify(tempServices) !== JSON.stringify(initialServices);

  return (
    <div className="space-y-4 relative">
      {isPending && (
        <div className="absolute inset-0 bg-background/50 rounded-lg flex items-center justify-center z-10 pointer-events-none">
          <div className="flex flex-col items-center gap-2">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
            <span className="text-sm text-muted-foreground">Updating filters...</span>
          </div>
        </div>
      )}
      <Card className="bg-card border-border">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center justify-between cursor-pointer uppercase tracking-wider text-sm" onClick={() => toggleSection('severity')}>
            <span>Vuln Severity</span>
            {expandedSections.severity ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </CardTitle>
        </CardHeader>
        {expandedSections.severity && (
          <CardContent className="p-0">
            <div className="space-y-3 px-4 py-2">
              {SEVERITY_OPTIONS.map((sev) => (
                <div key={sev} className="flex items-center gap-2">
                  <Checkbox
                    id={`severity-${sev}`}
                    checked={tempSeverities.includes(sev)}
                    onCheckedChange={() => handleSeverityToggle(sev)}
                    className="flex-shrink-0"
                  />
                  <Label
                    htmlFor={`severity-${sev}`}
                    className="cursor-pointer text-sm font-normal capitalize"
                  >
                    {sev}
                  </Label>
                </div>
              ))}
            </div>
          </CardContent>
        )}
      </Card>

      <Card className="bg-card border-border">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center justify-between cursor-pointer uppercase tracking-wider text-sm" onClick={() => toggleSection('countries')}>
            <span>Country</span>
            {expandedSections.countries ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </CardTitle>
        </CardHeader>
        {expandedSections.countries && (
          <CardContent className="p-0">
            <ScrollArea className="h-48">
              <div className="space-y-3 px-4 py-2">
                {(facets.countries ?? []).map((country) => (
                  <div key={country.code} className="flex items-center justify-between gap-2 pr-4">
                    <div className="flex items-center gap-2 min-w-0 flex-1">
                      <Checkbox
                        id={`country-${country.code}`}
                        checked={tempCountries.includes(country.code)}
                        onCheckedChange={() => handleCountryToggle(country.code)}
                        className="flex-shrink-0"
                      />
                      <Label
                        htmlFor={`country-${country.code}`}
                        className="cursor-pointer text-sm font-normal"
                        title={country.name}
                      >
                        {country.name}
                      </Label>
                    </div>
                    <span className="text-xs text-muted-foreground font-mono flex-shrink-0 whitespace-nowrap">
                      {country.count}
                    </span>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </CardContent>
        )}
      </Card>

      <Card className="bg-card border-border">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center justify-between cursor-pointer uppercase tracking-wider text-sm" onClick={() => toggleSection('asns')}>
            <span>ASN</span>
            {expandedSections.asns ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </CardTitle>
        </CardHeader>
        {expandedSections.asns && (
          <CardContent className="p-0">
            <ScrollArea className="h-48">
              <div className="space-y-3 px-4 py-2">
                {(facets.asns ?? []).map((asn) => {
                  const asnString = `AS${asn.code}`;
                  return (
                    <div key={asn.code} className="flex items-center justify-between gap-2 pr-4">
                      <div className="flex items-center gap-2 min-w-0 flex-1">
                        <Checkbox
                          id={`asn-${asnString}`}
                          checked={tempASNs.includes(asnString)}
                          onCheckedChange={() => handleASNToggle(asnString)}
                          className="flex-shrink-0"
                        />
                        <Label
                          htmlFor={`asn-${asnString}`}
                          className="cursor-pointer text-sm font-normal w-24"
                          title={asn.name}
                        >
                          {asn.name}
                        </Label>
                      </div>
                      <span className="text-xs text-muted-foreground font-mono flex-shrink-0 whitespace-nowrap">
                        {asn.count}
                      </span>
                    </div>
                  );
                })}
              </div>
            </ScrollArea>
          </CardContent>
        )}
      </Card>

      <Card className="bg-card border-border">
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center justify-between cursor-pointer uppercase tracking-wider text-sm" onClick={() => toggleSection('services')}>
            <span>Service</span>
            {expandedSections.services ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronRight className="h-4 w-4" />
            )}
          </CardTitle>
        </CardHeader>
        {expandedSections.services && (
          <CardContent className="p-0">
            <ScrollArea className="h-48">
              <div className="space-y-3 px-4 py-2">
                {Object.entries(facets.services || {}).map(([service, count]) => (
                  <div key={service} className="flex items-center justify-between gap-2 pr-4">
                    <div className="flex items-center gap-2 min-w-0 flex-1">
                      <Checkbox
                        id={`service-${service}`}
                        checked={tempServices.includes(service)}
                        onCheckedChange={() => handleServiceToggle(service)}
                        className="flex-shrink-0"
                      />
                      <Label
                        htmlFor={`service-${service}`}
                        className="cursor-pointer text-sm font-normal uppercase"
                        title={service}
                      >
                        {service}
                      </Label>
                    </div>
                    <span className="text-xs text-muted-foreground font-mono flex-shrink-0 whitespace-nowrap">
                      {count}
                    </span>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </CardContent>
        )}
      </Card>

      {hasChanges && (
        <Button
          onClick={applyFilters}
          disabled={isPending}
          className="w-full bg-primary hover:bg-primary/90"
        >
          {isPending ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              Applying Filters...
            </>
          ) : (
            'Apply Filters'
          )}
        </Button>
      )}
    </div>
  );
}
