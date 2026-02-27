# iBench - Internet Bandwidth Exporter

A Go application that measures internet speed using iperf3 and exports metrics to Grafana Cloud Prometheus.

**Features:**
- ✓ Automatic internet speed testing (download & upload)
- ✓ Metrics export to Grafana Cloud Prometheus
- ✓ Scheduled hourly execution via systemd timer
- ✓ YAML configuration support
- ✓ Graceful error handling
- ✓ Comprehensive logging

## Quick Start

### 1. Clone/Setup

```bash
git clone <repo>
cd ibenc
```

### 2. Install Dependencies

```bash
# Ensure you have iperf3 installed
sudo apt install iperf3  # Ubuntu/Debian
sudo dnf install iperf3  # Fedora/RHEL
brew install iperf3      # macOS
```

### 3. Configure

```bash
cp ibenc.yaml.example ibenc.yaml
nano ibenc.yaml
```

Update with your:
- Grafana Cloud Prometheus Instance ID
- API Token (with MetricsPublisher role)
- Location, ISP, and Package details

### 4. Build & Run

```bash
go build -o ibenc
./ibenc -config ibenc.yaml
```

Expected output:
```
2026/02/27 20:09:24 Test Results:
2026/02/27 20:09:24   Download: 88.60 Mbps
2026/02/27 20:09:24   Upload: 48.65 Mbps
2026/02/27 20:09:24   Latency: 0.00 ms
2026/02/27 20:09:24   Jitter: 0.00 ms
2026/02/27 20:09:24   Packet Loss: 0.00 %
2026/02/27 20:09:24 Sending metrics to https://prometheus-prod-01-eu-west-0.grafana.net/api/prom
2026/02/27 20:09:25 Metrics sent successfully!
```

### 5. Setup Hourly Automation (Optional)

See [SYSTEMD_SETUP.md](SYSTEMD_SETUP.md) for step-by-step instructions to run iBench every hour automatically.

Quick version:
```bash
go build -o ibenc
sudo cp ibenc /usr/local/bin/
sudo mkdir -p /etc/ibenc
sudo cp ibenc.yaml /etc/ibenc/
sudo cp ibenc.{service,timer} /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now ibenc.timer
```

## Configuration

See [CONFIG.md](CONFIG.md) for detailed configuration options.

Basic structure:
```yaml
prometheus:
  url: "https://prometheus-prod-01-eu-west-0.grafana.net/api/prom"
  username: "431589"              # Your instance ID
  password: "glc_..."             # Your API token

iperf3:
  server: "sgp.proof.ovh.net"
  port: 5201
  duration: 30                    # Test duration in seconds

metrics:
  location: "Gampaha, Sri Lanka"
  isp_name: "slt"
  package_name: "ftth-unlimited"
```

## Metrics Exported

| Metric | Description | Labels |
|--------|-------------|--------|
| `ibenc_download_speed_mbps` | Download speed | location, isp_name, package_name |
| `ibenc_upload_speed_mbps` | Upload speed | location, isp_name, package_name |
| `ibenc_latency_ms` | Network latency | location, isp_name, package_name |
| `ibenc_jitter_ms` | Network jitter | location, isp_name, package_name |
| `ibenc_packet_loss_percent` | Packet loss | location, isp_name, package_name |

## Architecture

```
┌─────────────────┐
│  Systemd Timer  │
│  (hourly)       │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│     iBench      │
│  ┌───────────┐  │
│  │ iperf3    │  │ ← Download & upload test
│  │ runner    │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼─────┐  │
│  │ Metrics   │  │ ← Format metrics
│  │ exporter  │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼──────────────────┐
│  │ Remote write protocol  │
│  │ (protobuf + snappy)    │
│  └─────┬──────────────────┘
└────────┼────────────────────┘
         │
         ↓
┌──────────────────────────────────┐
│  Grafana Cloud Prometheus        │
│  /api/prom/push endpoint         │
└──────────────────────────────────┘
         ↓
┌──────────────────────────────────┐
│  Grafana Cloud Dashboards        │
│  Visualize & Alert on metrics    │
└──────────────────────────────────┘
```

## File Structure

```
ibenc/
├── main.go                    # Entry point
├── go.mod                     # Dependencies
├── ibenc.yaml                 # Configuration (gitignored)
├── ibenc.yaml.example         # Example config
├── iperf3/
│   └── runner.go             # iperf3 test execution
├── metrics/
│   └── exporter.go           # Prometheus metrics formatting
├── remote/
│   ├── writer.go             # Remote write sender
│   └── writer_text.go        # Alternative text format
├── config/
│   └── config.go             # Configuration management
├── cmd/
│   ├── test-grafana/         # Testing tool
│   └── debug-metrics/        # Debug tool
├── ibenc.service             # Systemd service file
├── ibenc.timer               # Systemd timer
├── CONFIG.md                 # Configuration guide
├── SETUP.md                  # Setup instructions
└── SYSTEMD_SETUP.md         # Systemd automation guide
```

## Requirements

- Go 1.25.7+
- iperf3 installed
- Linux system with systemd (for automation)
- Grafana Cloud account with Prometheus
- Network connectivity to iperf3 server and Grafana Cloud

## Getting Grafana Cloud Credentials

1. Create free account: https://grafana.com/auth/sign-up/
2. Navigate to Prometheus in the sidebar
3. Click your instance (e.g., "prod-01-eu-west")
4. Copy **Instance ID** from Details tab
5. Go to Account Settings → API Tokens
6. Create new token with **MetricsPublisher** role
7. Copy the token (starts with `glc_`)

See [CONFIG.md](CONFIG.md) for detailed instructions.

## Troubleshooting

### iperf3 connection fails
- Check iperf3 is installed: `iperf3 --version`
- Verify server is reachable: `nc -zv sgp.proof.ovh.net 5201`
- Check firewall isn't blocking port 5201
- Try different iperf3 server

### Metrics not appearing in Grafana
- Verify Grafana Cloud credentials
- Check API token has MetricsPublisher role
- Review logs: `journalctl -u ibenc.service`
- Manually test: `./ibenc -config ibenc.yaml`

### Service not running
- Check timer status: `sudo systemctl status ibenc.timer`
- View logs: `sudo journalctl -u ibenc.service`
- Test manually: `sudo systemctl start ibenc.service`

See detailed troubleshooting in [CONFIG.md](CONFIG.md) and [SYSTEMD_SETUP.md](SYSTEMD_SETUP.md).

## Development

### Running Tests

```bash
# Test basic functionality
./ibenc -config ibenc.yaml

# Test with mock data (if available)
go run ./cmd/test-grafana/main.go -config ibenc.yaml
```

### Building Release Binary

```bash
go build -o ibenc
```

### Code Structure

- **iperf3/runner.go** - Executes iperf3 commands and parses JSON output
- **metrics/exporter.go** - Converts test results to Prometheus MetricFamily format
- **remote/writer.go** - Sends metrics using Prometheus remote write protocol (protobuf + snappy)
- **config/config.go** - Loads and validates YAML configuration

## License

[Your License Here]

## Contributing

Pull requests welcome! Please ensure:
- Code builds without errors
- Configuration examples are updated
- Tests pass
- Documentation is updated

## Support

For issues and questions:
1. Check [CONFIG.md](CONFIG.md) and [SYSTEMD_SETUP.md](SYSTEMD_SETUP.md)
2. Review application logs: `journalctl -u ibenc.service`
3. Run manual test: `./ibenc -config ibenc.yaml`
4. Check Grafana Cloud credentials and permissions
