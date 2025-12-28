"use client";

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { getHostByIP } from '../../../../lib/api';
import { Host } from '../../../../lib/types';
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
  Copy,
  Check,
} from 'lucide-react';

export default function HostDetailsPage() {
  const params = useParams();
  const slug = params?.slug as string;
  
  const [host, setHost] = useState<Host | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState<string | null>(null);

  useEffect(() => {
    const fetchHost = async () => {
      if (!slug) return;
      
      setLoading(true);
      setError(null);
      
      try {
        const data = await getHostByIP(slug);
        setHost(data);
      } catch (err) {
        console.error('Failed to fetch host:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch host details');
      } finally {
        setLoading(false);
      }
    };

    fetchHost();
  }, [slug]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const formatDateShort = (dateString: string) => {
    return new Date(dateString).toISOString().split('T')[0];
  };

  const copyToClipboard = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(id);
      setTimeout(() => setCopied(null), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background p-6 dark">
        <div className="max-w-6xl mx-auto">
          <div className="space-y-6">
            <Link href="/">
              <Button size="sm" className="gap-2">
                <ChevronLeft className="h-4 w-4" />
                Back to Search
              </Button>
            </Link>
            <div className="text-center py-12">
              <p className="text-muted-foreground">Loading host details...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !host) {
    return (
      <div className="min-h-screen bg-background p-6 dark">
        <div className="max-w-6xl mx-auto">
          <div className="space-y-6">
            <Link href="/">
              <Button size="sm" className="gap-2">
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
    <div className="min-h-screen bg-background p-6 dark">
      <div className="max-w-6xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <Link href="/">
            <Button size="sm" className="gap-2">
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
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => copyToClipboard(host.ip, 'ip')}
                  className="gap-2"
                >
                  {copied === 'ip' ? (
                    <>
                      <Check className="h-4 w-4" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4" />
                      Copy IP
                    </>
                  )}
                </Button>
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
                                          {values.join(', ')}
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
                                          {values.join(', ')}
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