'use client';

import { useState } from 'react';
import { Host } from '../lib/types';
import { HostCard } from './HostCard';
import { Button } from './ui/button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface HostResultsProps {
  hosts: Host[];
  totalCount: number;
  isLoading?: boolean;
}

const HOSTS_PER_PAGE = 10;

export function HostResults({ hosts, totalCount, isLoading }: HostResultsProps) {
  const [currentPage, setCurrentPage] = useState(1);

  const totalPages = Math.ceil(hosts.length / HOSTS_PER_PAGE);
  const startIndex = (currentPage - 1) * HOSTS_PER_PAGE;
  const endIndex = startIndex + HOSTS_PER_PAGE;
  const paginatedHosts = hosts.slice(startIndex, endIndex);

  // Reset to page 1 when filters change
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Reset page when hosts change
  if (currentPage > totalPages && totalPages > 0) {
    setCurrentPage(1);
  }

  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const showPages = 5;
    
    if (totalPages <= showPages) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      if (currentPage <= 3) {
        for (let i = 1; i <= 4; i++) {
          pages.push(i);
        }
        pages.push('...');
        pages.push(totalPages);
      } else if (currentPage >= totalPages - 2) {
        pages.push(1);
        pages.push('...');
        for (let i = totalPages - 3; i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        pages.push(1);
        pages.push('...');
        pages.push(currentPage - 1);
        pages.push(currentPage);
        pages.push(currentPage + 1);
        pages.push('...');
        pages.push(totalPages);
      }
    }
    
    return pages;
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          Showing <span className="text-foreground font-medium">{startIndex + 1}-{Math.min(endIndex, hosts.length)}</span> of{' '}
          <span className="text-foreground font-medium">{hosts.length}</span> results
          {totalCount !== hosts.length && (
            <span className="ml-1">
              ({totalCount} total hosts)
            </span>
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
            {paginatedHosts.map((host) => (
              <a href={`host/${host.ip}`} key={host.id} className="block">
                <HostCard key={host.id} host={host} />
              </a>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 pt-4">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handlePageChange(currentPage - 1)}
                disabled={currentPage === 1}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>

              {getPageNumbers().map((page, index) => (
                typeof page === 'number' ? (
                  <Button
                    key={index}
                    variant={currentPage === page ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => handlePageChange(page)}
                    className="min-w-[40px]"
                  >
                    {page}
                  </Button>
                ) : (
                  <span key={index} className="px-2 text-muted-foreground">
                    {page}
                  </span>
                )
              ))}

              <Button
                variant="outline"
                size="sm"
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage === totalPages}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
