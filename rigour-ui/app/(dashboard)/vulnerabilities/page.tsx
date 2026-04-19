import Link from 'next/link';
import { getVulnerabilitySummary } from '../../../lib/api';
import { VulnerabilityTemplateSummary, SeverityLevel } from '../../../lib/types';
import { Card, CardContent, CardHeader } from '../../../components/ui/card';
import { Badge } from '../../../components/ui/badge';
import { cn } from '../../../lib/utils';
import { SEVERITY_CLASSES } from '../../../lib/constants';
import { ShieldAlert, ExternalLink } from 'lucide-react';

const SEVERITY_ORDER: SeverityLevel[] = ['critical', 'high', 'medium', 'low', 'info'];

interface PageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function VulnerabilitiesPage({ searchParams: searchParamsPromise }: PageProps) {
  const searchParams = await searchParamsPromise;
  const severityFilter = typeof searchParams.severity === 'string' ? searchParams.severity : undefined;

  let vulns: VulnerabilityTemplateSummary[] = [];
  let error: string | null = null;

  try {
    vulns = await getVulnerabilitySummary(severityFilter);
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to fetch vulnerabilities';
  }

  // Count by severity for the summary bar
  const severityCounts: Record<string, number> = {};
  for (const v of vulns) {
    severityCounts[v.severity] = (severityCounts[v.severity] || 0) + 1;
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-red-500">Error: {error}</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <ShieldAlert className="h-6 w-6 text-muted-foreground" />
          <h1 className="text-2xl font-bold tracking-tight">Vulnerabilities</h1>
        </div>
        {severityFilter && (
          <Link
            href="/vulnerabilities"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            Clear filter
          </Link>
        )}
      </div>

      {/* Severity summary bar */}
      <div className="grid grid-cols-5 gap-3">
        {SEVERITY_ORDER.map((sev) => {
          const count = severityCounts[sev] || 0;
          const isActive = severityFilter === sev;
          return (
            <Link key={sev} href={isActive ? '/vulnerabilities' : `/vulnerabilities?severity=${sev}`}>
              <Card className={cn(
                'transition-all cursor-pointer hover:ring-1 hover:ring-ring',
                isActive && 'ring-2 ring-ring'
              )}>
                <CardContent className="p-4 text-center">
                  <div className="text-2xl font-bold">{count}</div>
                  <Badge variant="outline" className={cn('mt-1 uppercase text-xs', SEVERITY_CLASSES[sev])}>
                    {sev}
                  </Badge>
                </CardContent>
              </Card>
            </Link>
          );
        })}
      </div>

      {/* Vulnerability list */}
      {vulns.length === 0 ? (
        <div className="text-center py-12">
          <ShieldAlert className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <p className="text-muted-foreground">
            {severityFilter ? `No ${severityFilter} vulnerabilities found.` : 'No vulnerabilities found.'}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {vulns.map((vuln) => (
            <Card key={vuln.template_id} className="bg-card border-border">
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-1 min-w-0 flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <Badge variant="outline" className={cn('uppercase text-xs', SEVERITY_CLASSES[vuln.severity as SeverityLevel])}>
                        {vuln.severity}
                      </Badge>
                      <span className="font-semibold text-sm">{vuln.name}</span>
                      {vuln.cvss > 0 && (
                        <span className="text-xs text-muted-foreground font-mono">
                          CVSS {vuln.cvss.toFixed(1)}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground font-mono">{vuln.template_id}</p>
                  </div>
                  <Link
                    href={`/?filter=${encodeURIComponent(JSON.stringify({ 'vulns.template_id': vuln.template_id }))}`}
                    className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors whitespace-nowrap"
                  >
                    <span className="font-mono">{vuln.affected_host_count}</span> host{vuln.affected_host_count !== 1 && 's'}
                    <ExternalLink className="h-3 w-3" />
                  </Link>
                </div>
              </CardHeader>
              {(vuln.description || (vuln.cve_ids && vuln.cve_ids.length > 0)) && (
                <CardContent className="pt-0 pb-4">
                  {vuln.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2">{vuln.description}</p>
                  )}
                  {vuln.cve_ids && vuln.cve_ids.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 mt-2">
                      {vuln.cve_ids.map((cve) => (
                        <Badge key={cve} variant="outline" className="text-xs font-mono">
                          {cve}
                        </Badge>
                      ))}
                    </div>
                  )}
                </CardContent>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
