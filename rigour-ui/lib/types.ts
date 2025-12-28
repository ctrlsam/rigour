export interface Service {
  ip: string;
  port: number;
  protocol: string;
  tls: boolean;
  transport: string;
  last_scan: string;
  https?: {
    status: string;
    statusCode: number;
    responseHeaders: Record<string, string[]>;
  };
  http?: {
    status: string;
    statusCode: number;
    responseHeaders: Record<string, string[]>;
  };
  ssh?: {
    banner: string;
  };
}

export interface Host {
  id: string;
  ip: string;
  ip_int: number;
  asn: {
    number: number;
    organization: string;
    country: string;
  };
  location: {
    coordinates: [number, number];
    city: string;
    timezone: string;
  };
  first_seen: string;
  last_seen: string;
  services: Service[];
}

// Mock data for demonstration
const mockHosts: Host[] = [
  {
    id: '695108cef8f6bde74f741426',
    ip: '74.48.100.154',
    ip_int: 1244685466,
    asn: {
      number: 35916,
      organization: 'MULTA-ASN1',
      country: 'United States'
    },
    location: {
      coordinates: [-118.244, 34.0544],
      city: 'Los Angeles',
      timezone: 'America/Los_Angeles'
    },
    first_seen: '2025-12-28T10:29:10.02Z',
    last_seen: '2025-12-28T10:29:10.02Z',
    services: [
      {
        ip: '74.48.100.154',
        port: 443,
        protocol: 'https',
        tls: true,
        transport: 'tcp',
        last_scan: '2025-12-28T10:39:10.02Z',
        https: {
          status: '404 Not Found',
          statusCode: 404,
          responseHeaders: {
            'Content-Length': ['0'],
            'Date': ['Sun, 28 Dec 2025 10:39:09 GMT']
          }
        }
      }
    ]
  },
  {
    id: '695108cef8f6bde74f741427',
    ip: '203.0.113.42',
    ip_int: 3405803818,
    asn: {
      number: 3320,
      organization: 'Deutsche Telekom AG',
      country: 'Germany'
    },
    location: {
      coordinates: [13.4050, 52.5200],
      city: 'Berlin',
      timezone: 'Europe/Berlin'
    },
    first_seen: '2025-12-27T08:15:22.11Z',
    last_seen: '2025-12-28T14:22:33.45Z',
    services: [
      {
        ip: '203.0.113.42',
        port: 22,
        protocol: 'ssh',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T14:22:33.45Z',
        ssh: {
          banner: 'SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5'
        }
      },
      {
        ip: '203.0.113.42',
        port: 80,
        protocol: 'http',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T14:22:33.45Z',
        http: {
          status: '200 OK',
          statusCode: 200,
          responseHeaders: {
            'Server': ['nginx/1.18.0'],
            'Content-Type': ['text/html']
          }
        }
      }
    ]
  },
  {
    id: '695108cef8f6bde74f741428',
    ip: '198.51.100.77',
    ip_int: 3325256781,
    asn: {
      number: 2856,
      organization: 'British Telecommunications PLC',
      country: 'United Kingdom'
    },
    location: {
      coordinates: [-0.1278, 51.5074],
      city: 'London',
      timezone: 'Europe/London'
    },
    first_seen: '2025-12-26T19:45:10.88Z',
    last_seen: '2025-12-28T16:30:15.22Z',
    services: [
      {
        ip: '198.51.100.77',
        port: 554,
        protocol: 'rtsp',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T16:30:15.22Z'
      },
      {
        ip: '198.51.100.77',
        port: 8554,
        protocol: 'rtsp',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T16:30:15.22Z'
      }
    ]
  },
  {
    id: '695108cef8f6bde74f741429',
    ip: '172.217.14.206',
    ip_int: 2899908814,
    asn: {
      number: 15169,
      organization: 'Google LLC',
      country: 'United States'
    },
    location: {
      coordinates: [-122.0838, 37.4220],
      city: 'Mountain View',
      timezone: 'America/Los_Angeles'
    },
    first_seen: '2025-12-25T12:00:00.00Z',
    last_seen: '2025-12-28T09:15:45.33Z',
    services: [
      {
        ip: '172.217.14.206',
        port: 443,
        protocol: 'https',
        tls: true,
        transport: 'tcp',
        last_scan: '2025-12-28T09:15:45.33Z',
        https: {
          status: '200 OK',
          statusCode: 200,
          responseHeaders: {
            'Server': ['gws'],
            'Content-Type': ['text/html; charset=UTF-8']
          }
        }
      }
    ]
  },
  {
    id: '695108cef8f6bde74f74142a',
    ip: '185.220.101.15',
    ip_int: 3117752591,
    asn: {
      number: 24961,
      organization: 'myLoc managed IT AG',
      country: 'Netherlands'
    },
    location: {
      coordinates: [4.8945, 52.3702],
      city: 'Amsterdam',
      timezone: 'Europe/Amsterdam'
    },
    first_seen: '2025-12-27T20:30:10.55Z',
    last_seen: '2025-12-28T18:45:20.11Z',
    services: [
      {
        ip: '185.220.101.15',
        port: 9001,
        protocol: 'tor',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T18:45:20.11Z'
      },
      {
        ip: '185.220.101.15',
        port: 9030,
        protocol: 'tor-dir',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T18:45:20.11Z'
      }
    ]
  },
  {
    id: '695108cef8f6bde74f74142b',
    ip: '45.142.120.33',
    ip_int: 764328993,
    asn: {
      number: 205100,
      organization: 'Alibaba US Technology Co., Ltd.',
      country: 'Russia'
    },
    location: {
      coordinates: [37.6173, 55.7558],
      city: 'Moscow',
      timezone: 'Europe/Moscow'
    },
    first_seen: '2025-12-26T14:20:30.77Z',
    last_seen: '2025-12-28T11:35:42.88Z',
    services: [
      {
        ip: '45.142.120.33',
        port: 554,
        protocol: 'rtsp',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T11:35:42.88Z'
      },
      {
        ip: '45.142.120.33',
        port: 80,
        protocol: 'http',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T11:35:42.88Z',
        http: {
          status: '200 OK',
          statusCode: 200,
          responseHeaders: {
            'Server': ['DVR Webserver'],
            'Content-Type': ['text/html']
          }
        }
      }
    ]
  },
  {
    id: '695108cef8f6bde74f74142c',
    ip: '104.244.42.129',
    ip_int: 1760676481,
    asn: {
      number: 54113,
      organization: 'Fastly',
      country: 'United States'
    },
    location: {
      coordinates: [-73.9352, 40.7306],
      city: 'New York',
      timezone: 'America/New_York'
    },
    first_seen: '2025-12-28T06:10:15.22Z',
    last_seen: '2025-12-28T19:25:30.44Z',
    services: [
      {
        ip: '104.244.42.129',
        port: 1883,
        protocol: 'mqtt',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T19:25:30.44Z'
      },
      {
        ip: '104.244.42.129',
        port: 8883,
        protocol: 'mqtt',
        tls: true,
        transport: 'tcp',
        last_scan: '2025-12-28T19:25:30.44Z'
      }
    ]
  },
  {
    id: '695108cef8f6bde74f74142d',
    ip: '91.108.56.165',
    ip_int: 1528842917,
    asn: {
      number: 62041,
      organization: 'Telegram Messenger LLP',
      country: 'Netherlands'
    },
    location: {
      coordinates: [4.8945, 52.3702],
      city: 'Amsterdam',
      timezone: 'Europe/Amsterdam'
    },
    first_seen: '2025-12-24T15:40:20.66Z',
    last_seen: '2025-12-28T08:55:10.99Z',
    services: [
      {
        ip: '91.108.56.165',
        port: 5060,
        protocol: 'sip',
        tls: false,
        transport: 'tcp',
        last_scan: '2025-12-28T08:55:10.99Z'
      },
      {
        ip: '91.108.56.165',
        port: 5061,
        protocol: 'sip',
        tls: true,
        transport: 'tcp',
        last_scan: '2025-12-28T08:55:10.99Z'
      }
    ]
  }
];