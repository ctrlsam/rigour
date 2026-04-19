'use client';

import { useRouter, useSearchParams } from 'next/navigation';
import { Host } from '../lib/types';
import { HostCard } from './HostCard';
import { Button } from './ui/button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface HostResultsProps {
  hosts: Host[];
  totalCount: number;
  currentPage: number;
  nextPageToken?: string;
  isLoading?: boolean;
}

export function HostResults({ hosts, totalCount, currentPage, nextPageToken }: HostResultsProps) {
  const router = useRouter();
  const searchParams = useSearchParams();

  const handleNext = () => {
    const params = new URLSearchParams(searchParams.toString());
    params.set('page', String(currentPage + 1));
    params.set('page_token', nextPageToken!);
    router.push(`?${params.toString()}`);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handlePrev = () => {
    router.back();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const hasPrev = currentPage > 1;
  const hasNext = !!nextPageToken;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          <span className="text-foreground font-medium">{hosts.length}</span> results on page{' '}
          <span className="text-foreground font-medium">{currentPage}</span>
          {totalCount !== hosts.length && (
            <span className="ml-1">({totalCount} total hosts)</span>
          )}
        </div>
      </div>

      {hosts.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground">No hosts found matching your criteria.</p>
        </div>
      ) : (
        <>
          <div className="space-y-4">
            {hosts.map((host) => (
              <a href={`host/${host.ip}`} key={host.id} className="block">
                <HostCard key={host.id} host={host} />
              </a>
            ))}
          </div>

          {(hasPrev || hasNext) && (
            <div className="flex items-center justify-center gap-4 pt-4">
              <Button
                variant="outline"
                size="sm"
                onClick={handlePrev}
                disabled={!hasPrev}
              >
                <ChevronLeft className="h-4 w-4 mr-1" />
                Prev
              </Button>

              <span className="text-sm text-muted-foreground">Page {currentPage}</span>

              <Button
                variant="outline"
                size="sm"
                onClick={handleNext}
                disabled={!hasNext}
              >
                Next
                <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
