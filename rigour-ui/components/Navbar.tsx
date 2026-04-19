'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Github } from 'lucide-react';
import { cn } from '../lib/utils';

export function Navbar() {
  const pathname = usePathname();

  return (
    <nav className="border-b border-border bg-background sticky top-0 z-50">
      <div className="container mx-auto px-4 flex items-center justify-between h-14">
        <div className="flex items-center gap-8">
          <Link
            href="/"
            className="text-xl font-bold tracking-tighter text-foreground"
            style={{ fontFamily: 'Space Grotesk, monospace', letterSpacing: '-0.05em' }}
          >
            RIGOUR
          </Link>
          <div className="flex items-center gap-1 text-sm">
            <Link
              href="/"
              className={cn(
                'px-3 py-1.5 rounded-md transition-colors',
                pathname === '/'
                  ? 'bg-secondary text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Hosts
            </Link>
            <Link
              href="/vulnerabilities"
              className={cn(
                'px-3 py-1.5 rounded-md transition-colors',
                pathname.startsWith('/vulnerabilities')
                  ? 'bg-secondary text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Vulnerabilities
            </Link>
            <Link
              href="/campaigns"
              className={cn(
                'px-3 py-1.5 rounded-md transition-colors',
                pathname.startsWith('/campaigns')
                  ? 'bg-secondary text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Campaigns
            </Link>
          </div>
        </div>
        <a
          href="https://github.com/ctrlsam/rigour"
          target="_blank"
          rel="noopener noreferrer"
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          <Github className="h-5 w-5" />
        </a>
      </div>
    </nav>
  );
}
