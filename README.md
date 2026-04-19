# Rigour: An IoT Search Engine

[![GitHub License](https://img.shields.io/github/license/ctrlsam/rigour?style=flat-square)](LICENSE)
[![GitHub Issues](https://img.shields.io/github/issues/ctrlsam/rigour?style=flat-square)](https://github.com/ctrlsam/rigour/issues)
[![GitHub Stars](https://img.shields.io/github/stars/ctrlsam/rigour?style=flat-square)](https://github.com/ctrlsam/rigour)

Rigour is a comprehensive Internet of Things (IoT) scanning tool designed to discover, analyze, and report on devices connected to the internet. Rigour performs large-scale network scans to identify active hosts, retrieve service banners, and detect potential vulnerabilities. Rigour was inspired by Shodan.io, a popular IoT search engine. If you find this project useful, please consider starring the repository.

> [!WARNING]
> Rigour is intended for ethical use only. Always obtain permission before scanning networks and devices that you do not own. Use this tool responsibly and in compliance with all applicable laws and regulations.

## Get Started

Before you begin, ensure you have the necessary prerequisites installed on your system.

### Prerequisites

* [Docker](https://www.docker.com/get-started)
* [Docker Compose](https://docs.docker.com/compose/install/)
* [MaxMind Account](./docs/MAXMIND_SETUP.md)

### Installation Steps

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/ctrlsam/rigour.git
   cd rigour
   ```

### Run with Docker (Recommended)

1. **Configure Environment Variables**:
   Create a `.env` file under the root directory and set the required environment variables as per the instructions in `.env.example`.
   You will also want to set what IP range you want to scan. By default this is set as the ENTIRE internet so be careful!

   ```bash
    cp .env.example .env
    nano .env
    ```

2. **Run with Docker Compose**:
   Ensure you have Docker and Docker Compose installed. Then, run:

   ```bash
   docker compose up -d
   ```

3. **Access the UI**:
    Open your web browser and navigate to `http://localhost:3000` to access the Rigour web interface.

4. **Stop the Services**:
    To stop the services, run:
    ```bash
    docker compose down
    ```

## Architecture Overview

Rigour's architecture comprises several interconnected components that work in harmony to perform comprehensive network scanning and analysis.

### Components

#### Crawler

The Crawler is responsible for performing large-scale network scans using [Naabu](https://github.com/projectdiscovery/naabu) and fingerprinting the discovered devices with [Fingerprintx](https://github.com/praetorian-inc/fingerprintx). Results from this are published to a message bus for further processing by other services. The microservice design was chosen so that it would be easy to scale out the scanning infrastructure by adding more crawler instances as needed.

#### Persistence

The Persistence component consumes scan from the crawler and enriches them with other data sources such as GeoIP databases and vulnerability data. It then stores the enriched data in a NoSQL database. This allows for efficient querying and retrieval of scan data for analysis and reporting.

#### API

The API component provides a RESTful interface for accessing scan data stored in the database. It serves as the backend for the Rigour UI, enabling users to query and retrieve scan results.

#### User Interface

The Rigour UI provides an intuitive interface for viewing scan results. You can filter and search for specific devices, view detailed information about each device, and export scan results for further analysis. The app is build using Next.js and communicates with the API to fetch data.

<img src="./docs/ui.png" alt="Rigour UI Screenshot" width="500"/>

## REST API Documentation

### Hosts

#### `GET /api/hosts/search`

- **Description**: Search for hosts based on query parameters.
- **Query Parameters**:
    - `filter` (optional, string): A JSON-encoded MongoDB-style query object used to filter hosts. The server applies this object directly as a MongoDB $match stage.
    - `limit` (optional, integer): Maximum number of hosts to return. Defaults to 50. Minimum/invalid values default to 50. Maximum allowed value is 500.
    - `page_token` (optional, string): Opaque pagination token returned from a previous response next_page_token. Pass this token to retrieve the next page of results.

#### `GET /api/hosts/{ip}`

- **Description**: Retrieve a single host record by IP address.

#### `GET /api/facets`

- **Description**: Retrieve available facets for filtering search results.
- **Query Parameters**:
    - `filter` (optional, string): A JSON-encoded MongoDB-style query object to restrict the aggregation to a subset of hosts.

### Campaigns

Campaigns allow targeted scans of specific CIDR ranges with configurable port ranges and service checks. Progress is persisted to MongoDB — if the server restarts mid-scan, campaigns in `running` state are automatically resumed.

#### `GET /api/campaigns`

- **Description**: List all campaigns (id, name, status, progress).
- **Query Parameters**:
    - `status` (optional): Filter by status — `pending`, `running`, `paused`, `completed`, `failed`, or `cancelled`.

#### `GET /api/campaigns/{id}`

- **Description**: Get the current status and progress of a campaign.

#### `POST /api/admin/campaigns`

- **Description**: Create a new campaign.
- **Body** (JSON):
    - `name` (required, string)
    - `description` (optional, string)
    - `cidrs` (required if no `country_code`, array of strings): e.g. `["1.2.3.0/24", "10.0.0.0/8"]`
    - `country_code` (required if no `cidrs`, string): ISO 3166-1 alpha-2 country code. CIDRs are resolved automatically from the MaxMind database at scan start.
    - `ports` (optional, string): Explicit port list — e.g. `"80,443,8080-8090"`. Mutually exclusive with `top_ports`.
    - `top_ports` (optional, string): Naabu preset — `"100"`, `"1000"`, or `"full"`. Defaults to `"1000"`.
    - `service_filter` (optional, array of strings): Limit fingerprinting to specific protocols — e.g. `["http","ssh"]`. Empty means all 43 supported protocols.
    - `scan_type` (optional, string): `"s"` (SYN, default) or `"c"` (connect).
    - `rate` (optional, integer): Naabu packet rate. Defaults to `5000`.
    - `retries` (optional, integer): Naabu retry count. Defaults to `1`.
    - `timeout` (optional, integer): Fingerprint timeout in milliseconds. Defaults to `1000`.
    - `fast_mode` (optional, boolean): Only probe default ports per protocol. Defaults to `false`.
    - `use_udp` (optional, boolean): Enable UDP plugins. Defaults to `false`.
    - `run_vuln_scan` (optional, boolean): Run Nuclei vulnerability scanning on each discovered service. Requires the server to be started with `--vuln-scan`. Defaults to `false`.
    - `vuln_scan_templates` (optional, array of strings): Nuclei template tags to use — e.g. `["cve", "rce"]`. Defaults to `["cve", "rce", "default-login", "exposure"]`.

#### `GET /api/admin/campaigns`

- **Description**: List all campaigns with full detail.
- **Query Parameters**:
    - `status` (optional): Filter by campaign status.

#### `GET /api/admin/campaigns/{id}`

- **Description**: Get full campaign detail including configuration and progress.

#### `POST /api/admin/campaigns/{id}/start`

- **Description**: Start a `pending` or resume a `paused` campaign.

#### `POST /api/admin/campaigns/{id}/pause`

- **Description**: Pause a running campaign after the current CIDR block finishes. Progress is preserved.

#### `POST /api/admin/campaigns/{id}/resume`

- **Description**: Resume a paused campaign from where it left off.

#### `POST /api/admin/campaigns/{id}/cancel`

- **Description**: Cancel a running campaign immediately.

#### `DELETE /api/admin/campaigns/{id}`

- **Description**: Delete a campaign. The campaign must not be in `running` state.

### Vulnerabilities

Vulnerability findings are stored per-service and keyed by `(ip, port, template_id)` — re-running a scan updates the record but preserves the original discovery timestamp.

#### `GET /api/hosts/{ip}/vulnerabilities`

- **Description**: List all vulnerabilities found on a specific host.

#### `GET /api/vulnerabilities`

- **Description**: List vulnerability findings, optionally scoped to a campaign.
- **Query Parameters**:
    - `campaign_id` (optional, string): Scope results to a specific campaign. When omitted, returns vulnerabilities across all campaigns.
    - `severity` (optional, string): Filter by severity — `critical`, `high`, `medium`, `low`, or `info`.

#### `GET /api/vulnerabilities/summary`

- **Description**: Returns vulnerabilities grouped by template, sorted by severity (critical first). Each entry includes the template name, severity, CVSS score, CVE IDs, and the number of affected hosts.
- **Query Parameters**:
    - `severity` (optional, string): Filter by severity — `critical`, `high`, `medium`, `low`, or `info`.

## Vulnerability Scanning

Rigour integrates [Nuclei](https://github.com/projectdiscovery/nuclei) for automated vulnerability scanning. When enabled, each discovered service is scanned against a curated set of Nuclei templates after fingerprinting completes. Scanning is opt-in per campaign and runs asynchronously — it never blocks the port discovery or fingerprinting pipeline.

### Enable at server startup

Pass `--vuln-scan` when starting the server. Nuclei templates are downloaded automatically on first run (stored at `~/nuclei-templates`). You can also point to a custom template directory with `--nuclei-templates`:

```bash
./rigour-server \
  --mongo-uri mongodb://localhost:27017 \
  --geoip-path /path/to/geoip \
  --vuln-scan \
  --nuclei-templates /path/to/nuclei-templates \  # optional; defaults to ~/nuclei-templates
  --nuclei-concurrency 5                          # optional; max concurrent scans
```

With Docker Compose, add the flags to the `command` block in `docker-compose.yml`:

```yaml
command:
  - "--mongo-uri"
  - "mongodb://mongo:27017"
  - "--geoip-path"
  - "/data/geoip"
  - "--vuln-scan"
```

### Enable per campaign

Set `run_vuln_scan: true` when creating a campaign:

```bash
curl -X POST http://localhost:8080/api/admin/campaigns \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Vuln Scan",
    "cidrs": ["192.168.1.0/24"],
    "ports": "80,443,8080",
    "run_vuln_scan": true,
    "vuln_scan_templates": ["cve", "default-login"]
  }'
```

Query findings after the scan:

```bash
# All findings for a campaign
curl "http://localhost:8080/api/vulnerabilities?campaign_id=<id>&severity=critical"

# Findings on a specific host
curl "http://localhost:8080/api/hosts/1.2.3.4/vulnerabilities"
```

## Development

### Benchmarking

The fingerprint pipeline has a benchmark suite that measures scan performance without requiring Docker or external network access. Use it to measure the impact of changes to scan algorithms.

```bash
go test ./rigour/pkg/crawler/fingerprint/... -bench=. -benchtime=5s -benchmem
```

Four benchmarks are included:

| Benchmark | What it measures |
|-----------|-----------------|
| `BenchmarkFingerprintUnknown` | Worst-case slow-lane: all 28+ TCP plugins tried against an unrecognised banner |
| `BenchmarkFingerprintUnknownFastMode` | Fast-lane cost when no plugin's default port matches |
| `BenchmarkFingerprintHTTPSlow` | Full slow-lane scan that successfully identifies HTTP |
| `BenchmarkFingerprintRedisSlow` | Full slow-lane scan that successfully identifies Redis |

Each benchmark spins up a local `net.Listen(":0")` mock server — no Docker, no external network required.

To compare performance before and after a change, use [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
# Capture baseline
go test ./rigour/pkg/crawler/fingerprint/... -bench=. -benchtime=5s -count=5 > before.txt

# Make your changes, then capture again
go test ./rigour/pkg/crawler/fingerprint/... -bench=. -benchtime=5s -count=5 > after.txt

benchstat before.txt after.txt
```

## Performance

> [!IMPORTANT]
> SYN scanning (the default) requires raw socket access, which means the server must run with **root / sudo** privileges. Without elevated privileges Rigour falls back to Connect scanning, which is significantly slower. For best performance, always run the server with `sudo` (or as root inside Docker, which is the default).

## Supported Systems

Rigour is designed to be platform-independent and can run on any system that supports Docker. Currently Rigour is primarily tested on Linux-based systems which perform best. I've also tested it on Apple Silicon Macs but performance is not as good for port discovery, will need to address this in future.

## Future Features

- [ ] Authentication — API key or JWT auth for admin campaign endpoints
- [ ] Distributed Crawler Support (e.g., Kubernetes & CIDR scheduling)

## Acknowledgements

We would like to thank the open-source community for their contributions and support in developing Rigour.

Special thanks to the creators of [Fingerprintx](https://github.com/praetorian-inc/fingerprintx) and [Naabu](https://github.com/projectdiscovery/naabu) for their invaluable tools and resources.
