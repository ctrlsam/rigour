'use client';

import { Search, Github } from 'lucide-react';
import { Input } from './ui/input';
import { Button } from './ui/button';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

interface SearchHeaderProps {
  initialQuery?: string;
}

export function SearchHeader({ initialQuery = '' }: SearchHeaderProps) {
  const [query, setQuery] = useState(initialQuery);
  const router = useRouter();

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
    const params = new URLSearchParams();
    
    if (query) {
      const parsedParams = parseQuery(query);
      
      if (parsedParams) {
        // Use parsed filter as query parameter
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
      <div className="flex items-center justify-between border-b border-border pb-6">
        <h1 className="tracking-tighter text-foreground" style={{ fontFamily: 'Space Grotesk, monospace', fontSize: '3rem', fontWeight: 700, letterSpacing: '-0.05em' }}>
          RIGOUR
        </h1>
        <a 
          href="https://github.com/ctrlsam/rigour" 
          target="_blank" 
          rel="noopener noreferrer"
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          <Github className="h-6 w-6" />
        </a>
      </div>
      
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