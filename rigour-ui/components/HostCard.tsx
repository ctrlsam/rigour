import { Host } from '../lib/types';
import { Card, CardContent, CardHeader } from './ui/card';
import { Badge } from './ui/badge';
import { Globe, Network, Server, Clock, MapPin } from 'lucide-react';

interface HostCardProps {
  host: Host;
  onClick?: () => void;
}

export function HostCard({ host, onClick }: HostCardProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toISOString().split('T')[0];
  };

  return (
    <Card 
      className="bg-card border-border hover:border-primary/50 transition-colors cursor-pointer"
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <h3 className="text-xl font-mono tracking-tight text-blue-400">{host.ip}</h3>
            <div className="grid grid-cols-2 gap-x-6 gap-y-1 text-xs text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <Globe className="h-3 w-3" />
                <span>{host.asn.country}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <MapPin className="h-3 w-3" />
                <span>{host.location.city}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <Network className="h-3 w-3" />
                <span className="font-mono">AS{host.asn.number}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <Clock className="h-3 w-3" />
                <span>{formatDate(host.last_seen)}</span>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">{host.asn.organization}</p>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Server className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground uppercase tracking-wider">Services</span>
          </div>
          <div className="space-y-2">
            {host.services.map((service, idx) => (
              <div key={idx} className="flex items-center gap-3 text-sm">
                <Badge variant="outline" className="font-mono px-2 py-0.5">
                  {service.port}
                </Badge>
                <span className="uppercase text-xs tracking-wide">{service.protocol}</span>
                {service.tls && (
                  <Badge variant="secondary" className="text-xs px-2 py-0.5">
                    TLS
                  </Badge>
                )}
                <span className="text-xs text-muted-foreground uppercase">{service.transport}</span>
              </div>
            ))}
          </div>
        </div>

        {host.services.some(s => s.https || s.http || s.ssh) && (
          <div>
            <div className="text-xs text-muted-foreground uppercase tracking-wider mb-2">Details</div>
            <div className="space-y-1 text-xs font-mono">
              {host.services.map((service, idx) => {
                if (service.https) {
                  return (
                    <div key={idx} className="text-muted-foreground">
                      HTTPS: {service.https.statusCode} {service.https.status}
                    </div>
                  );
                }
                if (service.http) {
                  return (
                    <div key={idx} className="text-muted-foreground">
                      HTTP: {service.http.statusCode} {service.http.status}
                    </div>
                  );
                }
                if (service.ssh) {
                  return (
                    <div key={idx} className="text-muted-foreground truncate">
                      SSH: {service.ssh.banner}
                    </div>
                  );
                }
                return null;
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}