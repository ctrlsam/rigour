'use client';

import { Search } from 'lucide-react';
import { Input } from './ui/input';
import { Button } from './ui/button';
import { useRouter, useSearchParams } from 'next/navigation';
import { useState } from 'react';

interface SearchHeaderProps {
  initialQuery?: string;
}

export function SearchHeader({ initialQuery = '' }: SearchHeaderProps) {
  const [query, setQuery] = useState(initialQuery);
  const router = useRouter();
  const searchParams = useSearchParams();

  // Parse query syntax like "field: value" into filter parameters
  const parseQuery = (queryString: string): Record<string, string> | null => {
    // Look for pattern like "field: value" or "field:value"
    const match = queryString.match(/(\w+(?:\.\w+)*)\s*:\s*(.+)/);

    if (match) {
      const [, field, value] = match;
      return { [field]: value.trim() };
    }

    return null;
  };

  const handleSearch = () => {
    const params = new URLSearchParams(searchParams.toString());

    // Reset pagination and previous search filters
    params.delete('page');
    params.delete('page_token');
    params.delete('filter');
    params.delete('query');

    // Facet params (countries, asns, services) are preserved from searchParams

    if (query) {
      params.set('query', query);
      const parsedParams = parseQuery(query);
      if (parsedParams) {
        for (const [key, value] of Object.entries(parsedParams)) {
          params.append('filter', JSON.stringify({ [key]: value }));
        }
      }
    }

    router.push(`?${params.toString()}`);
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  return (
    <header className="space-y-6">
      <div className="space-y-2">
        <p className="text-muted-foreground uppercase tracking-wider text-sm">
          Internet-Connected Device Intelligence Platform
        </p>
        <p className="text-xs text-muted-foreground">
          Query using field syntax: <code className="bg-secondary px-1 rounded">field: value</code> (e.g., <code className="bg-secondary px-1 rounded">services.protocol: ssh</code>)
        </p>
      </div>

      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
          <Input
            type="text"
            placeholder="e.g. services.protocol: ssh"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyPress={handleKeyPress}
            className="pl-12 h-14 bg-secondary border-border text-white"
          />
        </div>
        <Button
          onClick={handleSearch}
          className="h-14 px-6 bg-primary hover:bg-primary/90"
        >
          Search
        </Button>
      </div>
    </header>
  );
}
