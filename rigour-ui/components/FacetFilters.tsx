'use client';

import { useState, useTransition } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Checkbox } from './ui/checkbox';
import { Label } from './ui/label';
import { ScrollArea } from './ui/scroll-area';
import { Button } from './ui/button';
import { ChevronDown, ChevronRight, Loader2 } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { FacetCounts } from '../lib/api';

interface FacetFiltersProps {
  countries: string[];
  asns: string[];
  services: string[];
  selectedCountries: string[];
  selectedASNs: string[];
  selectedServices: string[];
  facets: FacetCounts;
}

export function FacetFilters({
  countries,
  asns,
  services,
  selectedCountries: initialCountries,
  selectedASNs: initialASNs,
  selectedServices: initialServices,
  facets,
}: FacetFiltersProps) {
  const [expandedSections, setExpandedSections] = useState({
    countries: true,
    asns: true,
    services: true,
  });
  // Local state for selections before applying
  const [tempCountries, setTempCountries] = useState(initialCountries);
  const [tempASNs, setTempASNs] = useState(initialASNs);
  const [tempServices, setTempServices] = useState(initialServices);
  
  const [isPending, startTransition] = useTransition();
  const router = useRouter();

  const toggleSection = (section: keyof typeof expandedSections) => {
    setExpandedSections(prev => ({
      ...prev,
      [section]: !prev[section],
    }));
  };

  const applyFilters = () => {
    startTransition(() => {
      const params = new URLSearchParams();
      
      if (tempCountries.length > 0) {
        params.append('countries', tempCountries.join(','));
      }
      
      if (tempASNs.length > 0) {
        params.append('asns', tempASNs.join(','));
      }
      
      if (tempServices.length > 0) {
        params.append('services', tempServices.join(','));
      }
      
      router.push(`?${params.toString()}`);
    });
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
    JSON.stringify(tempCountries) !== JSON.stringify(initialCountries) ||
    JSON.stringify(tempASNs) !== JSON.stringify(initialASNs) ||
    JSON.stringify(tempServices) !== JSON.stringify(initialServices);

  const getCountryCount = (country: string) =>
    facets.countries?.[country] || 0;

  const getASNCount = (asn: string) => {
    const asnNum = asn.replace('AS', '');
    const key = Object.keys(facets.asns || {}).find(k =>
      k.startsWith(asnNum + '-')
    );
    return key ? facets.asns![key] : 0;
  };

  const getServiceCount = (service: string) =>
    facets.services?.[service] || 0;

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
          <CardContent>
            <ScrollArea className="h-48">
              <div className="space-y-3">
                {countries.map((country) => (
                  <div key={country} className="flex items-center justify-between space-x-2">
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id={`country-${country}`}
                        checked={tempCountries.includes(country)}
                        onCheckedChange={() => handleCountryToggle(country)}
                      />
                      <Label
                        htmlFor={`country-${country}`}
                        className="cursor-pointer text-sm font-normal"
                      >
                        {country}
                      </Label>
                    </div>
                    <span className="text-xs text-muted-foreground font-mono">
                      {getCountryCount(country)}
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
          <CardContent>
            <ScrollArea className="h-48">
              <div className="space-y-3">
                {asns.map((asn) => (
                  <div key={asn} className="flex items-center justify-between space-x-2">
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id={`asn-${asn}`}
                        checked={tempASNs.includes(asn)}
                        onCheckedChange={() => handleASNToggle(asn)}
                      />
                      <Label
                        htmlFor={`asn-${asn}`}
                        className="cursor-pointer text-sm font-normal font-mono"
                      >
                        {asn}
                      </Label>
                    </div>
                    <span className="text-xs text-muted-foreground font-mono">
                      {getASNCount(asn)}
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
          <CardContent>
            <ScrollArea className="h-48">
              <div className="space-y-3">
                {services.map((service) => (
                  <div key={service} className="flex items-center justify-between space-x-2">
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id={`service-${service}`}
                        checked={tempServices.includes(service)}
                        onCheckedChange={() => handleServiceToggle(service)}
                      />
                      <Label
                        htmlFor={`service-${service}`}
                        className="cursor-pointer text-sm font-normal uppercase"
                      >
                        {service}
                      </Label>
                    </div>
                    <span className="text-xs text-muted-foreground font-mono">
                      {getServiceCount(service)}
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