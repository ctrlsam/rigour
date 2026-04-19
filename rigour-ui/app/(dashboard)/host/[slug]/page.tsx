import Link from 'next/link';
import { getHostByIP, getVulnsByIP } from '../../../../lib/api';
import { formatDate, formatDateShort } from '../../../../lib/utils';
import { Host, Vulnerability, SeverityLevel } from '../../../../lib/types';
import { Card, CardContent, CardHeader } from '../../../../components/ui/card';
import { Badge } from '../../../../components/ui/badge';
import { Button } from '../../../../components/ui/button';
import {
  Globe,
  Network,
  Server,
  Clock,
  MapPin,
  ChevronLeft,
  ExternalLink,
  AlertCircle,
  CheckCircle,
  Wifi,
  Shield,
  ShieldAlert,
} from 'lucide-react';

const SEVERITY_CLASSES: Record<SeverityLevel, string> = {
  critical: 'bg-red-500/20 text-red-400 border-red-500/30',
  high:     'bg-orange-500/20 text-orange-400 border-orange-500/30',
  medium:   'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  low:      'bg-blue-500/20 text-blue-400 border-blue-500/30',
  info:     'bg-gray-500/20 text-gray-400 border-gray-500/30',
};

interface Params {
  slug: string;
}

export default async function HostDetailsPage({ 
  params 
}: { 
  params: Promise<Params> 
}) {
  const { slug } = await params;
  let host: Host | null = null;
  let vulns: Vulnerability[] = [];
  let error: string | null = null;

  try {
    [host, vulns] = await Promise.all([getHostByIP(slug), getVulnsByIP(slug)]);
  } catch (err) {
    console.error('Failed to fetch host:', err);
    error = err instanceof Error ? err.message : 'Failed to fetch host details';
  }

  if (error || !host) {
    return (
      <div className="p-6">
        <div className="max-w-6xl mx-auto">
          <div className="space-y-6">
            <Link href="/">
              <Button size="sm" variant="outline" className="gap-2">
                <ChevronLeft className="h-4 w-4" />
                Back to Search
              </Button>
            </Link>
            <div className="text-center py-12">
              <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
              <p className="text-muted-foreground">
                {error || 'Host not found'}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="max-w-6xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <Link href="/">
            <Button size="sm" variant="outline" className="gap-2">
              <ChevronLeft className="h-4 w-4" />
              Back to Search
            </Button>
          </Link>
          <div className="text-sm text-muted-foreground">
            ID: <code className="bg-muted px-2 py-1 rounded text-xs">{host.id}</code>
          </div>
        </div>

        {/* Host Summary */}
        <Card className="bg-card border-border">
          <CardHeader className="pb-4">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h1 className="text-4xl font-mono tracking-tight text-blue-400 mb-2">
                    {host.ip}
                  </h1>
                  <p className="text-lg text-muted-foreground">{host.asn.organization}</p>
                </div>
              </div>

              {/* Key Information Grid */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t border-border">
                {/* Country */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                    <Globe className="h-4 w-4" />
                    Country
                  </div>
                  <p className="text-sm font-medium">{host.asn.country}</p>
                </div>

                {/* City */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                    <MapPin className="h-4 w-4" />
                    City
                  </div>
                  <p className="text-sm font-medium">{host.location.city}</p>
                </div>

                {/* ASN */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                    <Network className="h-4 w-4" />
                    ASN
                  </div>
                  <p className="text-sm font-mono font-medium">AS{host.asn.number}</p>
                </div>

                {/* Timezone */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                    <Clock className="h-4 w-4" />
                    Timezone
                  </div>
                  <p className="text-sm font-medium">{host.location.timezone}</p>
                </div>
              </div>
            </div>
          </CardHeader>
        </Card>

        {/* Timeline Information */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Timeline
            </h3>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <div className="text-xs text-muted-foreground uppercase tracking-wider font-medium">
                  First Seen
                </div>
                <p className="text-sm">{formatDate(host.first_seen)}</p>
                <p className="text-xs text-muted-foreground">
                  {formatDateShort(host.first_seen)}
                </p>
              </div>
              <div className="space-y-2">
                <div className="text-xs text-muted-foreground uppercase tracking-wider font-medium">
                  Last Seen
                </div>
                <p className="text-sm">{formatDate(host.last_seen)}</p>
                <p className="text-xs text-muted-foreground">
                  {formatDateShort(host.last_seen)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Geolocation Details */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <MapPin className="h-5 w-5" />
              Geolocation
            </h3>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <div className="text-xs text-muted-foreground uppercase tracking-wider font-medium">
                  Coordinates
                </div>
                <p className="text-sm font-mono">
                  {host.location.coordinates[0].toFixed(4)}, {host.location.coordinates[1].toFixed(4)}
                </p>
                <a
                  href={`https://maps.google.com/?q=${host.location.coordinates[1]},${host.location.coordinates[0]}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs text-blue-400 hover:text-blue-300 flex items-center gap-1 mt-2"
                >
                  View on Google Maps
                  <ExternalLink className="h-3 w-3" />
                </a>
              </div>
              <div className="space-y-2">
                <div className="text-xs text-muted-foreground uppercase tracking-wider font-medium">
                  ASN Information
                </div>
                <div className="space-y-1">
                  <p className="text-sm">
                    <span className="text-muted-foreground">Number:</span>{' '}
                    <span className="font-mono">AS{host.asn.number}</span>
                  </p>
                  <p className="text-sm">
                    <span className="text-muted-foreground">Org:</span>{' '}
                    <span className="font-medium">{host.asn.organization}</span>
                  </p>
                  <p className="text-sm">
                    <span className="text-muted-foreground">Country:</span>{' '}
                    <span className="font-medium">{host.asn.country}</span>
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Services */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <Server className="h-5 w-5" />
              Services ({host.services.length})
            </h3>
          </CardHeader>
          <CardContent className="space-y-4">
            {host.services.length === 0 ? (
              <p className="text-muted-foreground text-sm">No services discovered</p>
            ) : (
              <div className="space-y-4">
                {host.services.map((service, idx) => (
                  <div
                    key={idx}
                    className="p-4 border border-border rounded-lg space-y-3 hover:border-primary/50 transition-colors"
                  >
                    {/* Port and Protocol */}
                    <div className="flex items-center justify-between gap-2 flex-wrap">
                      <div className="flex items-center gap-3">
                        <Badge variant="outline" className="font-mono text-base px-3 py-1">
                          {service.port}
                        </Badge>
                        <span className="uppercase text-sm font-semibold tracking-wide">
                          {service.protocol}
                        </span>
                        {service.tls && (
                          <Badge variant="secondary" className="gap-1 px-2 py-1">
                            <Shield className="h-3 w-3" />
                            TLS
                          </Badge>
                        )}
                        {(service.protocol === 'http' || service.protocol === 'https') && (
                          <a
                            href={`${service.tls || service.protocol === 'https' ? 'https' : 'http'}://${host.ip}:${service.port}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300 transition-colors"
                          >
                            <ExternalLink className="h-3.5 w-3.5" />
                            Open
                          </a>
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <Wifi className="h-3 w-3" />
                        {service.transport.toUpperCase()}
                      </div>
                    </div>

                    {/* Last Scan */}
                    <div className="text-xs text-muted-foreground">
                      <span>Last scanned: {formatDateShort(service.last_scan)}</span>
                    </div>

                    {/* Service Details */}
                    {(service.https || service.http || service.ssh) && (
                      <div className="pt-2 border-t border-border space-y-2">
                        {service.https && (
                          <div className="space-y-2">
                            <div className="flex items-center gap-2">
                              <CheckCircle className="h-4 w-4 text-green-500" />
                              <span className="text-sm font-semibold">HTTPS</span>
                            </div>
                            <div className="ml-6 space-y-1 text-xs text-muted-foreground">
                              <p>
                                <span className="font-mono">Status Code: </span>
                                <span className="text-foreground font-medium">
                                  {service.https.statusCode}
                                </span>
                              </p>
                              <p>
                                <span className="font-mono">Status: </span>
                                <span className="text-foreground">{service.https.status}</span>
                              </p>
                              {Object.keys(service.https.responseHeaders).length > 0 && (
                                <div className="pt-2 space-y-1">
                                  <p className="font-semibold text-foreground">Headers:</p>
                                  <div className="pl-2 space-y-1 font-mono text-xs">
                                    {Object.entries(service.https.responseHeaders).map(
                                      ([key, values]) => (
                                        <p key={key}>
                                          <span className="text-blue-400">{key}:</span>{' '}
                                          {(values as string[]).join(', ')}
                                        </p>
                                      )
                                    )}
                                  </div>
                                </div>
                              )}
                            </div>
                          </div>
                        )}

                        {service.http && (
                          <div className="space-y-2">
                            <div className="flex items-center gap-2">
                              <CheckCircle className="h-4 w-4 text-green-500" />
                              <span className="text-sm font-semibold">HTTP</span>
                            </div>
                            <div className="ml-6 space-y-1 text-xs text-muted-foreground">
                              <p>
                                <span className="font-mono">Status Code: </span>
                                <span className="text-foreground font-medium">
                                  {service.http.statusCode}
                                </span>
                              </p>
                              <p>
                                <span className="font-mono">Status: </span>
                                <span className="text-foreground">{service.http.status}</span>
                              </p>
                              {Object.keys(service.http.responseHeaders).length > 0 && (
                                <div className="pt-2 space-y-1">
                                  <p className="font-semibold text-foreground">Headers:</p>
                                  <div className="pl-2 space-y-1 font-mono text-xs">
                                    {Object.entries(service.http.responseHeaders).map(
                                      ([key, values]) => (
                                        <p key={key}>
                                          <span className="text-blue-400">{key}:</span>{' '}
                                          {(values as string[]).join(', ')}
                                        </p>
                                      )
                                    )}
                                  </div>
                                </div>
                              )}
                            </div>
                          </div>
                        )}

                        {service.ssh && (
                          <div className="space-y-2">
                            <div className="flex items-center gap-2">
                              <CheckCircle className="h-4 w-4 text-green-500" />
                              <span className="text-sm font-semibold">SSH</span>
                            </div>
                            <div className="ml-6 space-y-1 text-xs text-muted-foreground">
                              <p className="font-mono break-all">{service.ssh.banner}</p>
                            </div>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Vulnerabilities */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h3 className="text-lg font-semibold flex items-center gap-2">
              <ShieldAlert className="h-5 w-5" />
              Vulnerabilities ({vulns.length})
            </h3>
          </CardHeader>
          <CardContent>
            {vulns.length === 0 ? (
              <p className="text-muted-foreground text-sm">No vulnerabilities found</p>
            ) : (
              <div className="space-y-3">
                {vulns.map(vuln => (
                  <div key={vuln.id} className="p-4 border border-border rounded-lg space-y-2 hover:border-primary/50 transition-colors">
                    <div className="flex items-center gap-3 flex-wrap">
                      <Badge variant="outline" className={`text-xs px-2 py-0.5 uppercase font-semibold ${SEVERITY_CLASSES[vuln.severity]}`}>
                        {vuln.severity}
                      </Badge>
                      <span className="text-sm font-medium">{vuln.name}</span>
                      <span className="text-xs text-muted-foreground font-mono">:{vuln.port}/{vuln.protocol}</span>
                      {vuln.cvss !== undefined && vuln.cvss > 0 && (
                        <span className="text-xs text-muted-foreground">CVSS {vuln.cvss.toFixed(1)}</span>
                      )}
                    </div>
                    {vuln.description && (
                      <p className="text-xs text-muted-foreground">{vuln.description}</p>
                    )}
                    <div className="flex items-center gap-3 flex-wrap">
                      {vuln.cve_ids && vuln.cve_ids.map(cve => (
                        <span key={cve} className="text-xs font-mono bg-red-950/40 text-red-300 border border-red-800/40 rounded px-1.5 py-0.5">
                          {cve}
                        </span>
                      ))}
                      <span className="text-xs text-muted-foreground ml-auto">
                        Found {formatDateShort(vuln.discovered_at)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Raw Data Section */}
        <Card className="bg-card border-border">
          <CardHeader>
            <h3 className="text-lg font-semibold">Raw Data</h3>
          </CardHeader>
          <CardContent>
            <div className="bg-muted p-4 rounded-lg overflow-auto max-h-96">
              <pre className="text-xs font-mono text-muted-foreground">
                {JSON.stringify(host, null, 2)}
              </pre>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
