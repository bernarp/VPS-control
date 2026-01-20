# VPS Control API

VPS management API built with Go and Gin framework.

## CI/CD Pipeline

GitHub Actions workflow with 5 stages:

1. Setup - Generate Swagger docs and upload as artifact
2. Lint - golangci-lint + govulncheck
3. Test - Unit tests with race detector
4. Build - Static binary for Linux AMD64
5. Deploy - SCP binary to VPS + systemctl restart

### Required GitHub Secrets

| Secret | Description | Example |
|--------|-------------|---------|
| HOST | VPS hostname | `192.168.1.100` |
| USERNAME | SSH user | `deploy` |
| SSH_PRIVATE_KEY | Private key for authentication | `-----BEGIN OPENSSH...` |
| DEPLOY_PATH | Deployment directory | `/opt/apps/vps-control` |
| SERVICE_NAME | systemd service name | `vps-api.service` |

### Tests and Checks

- **golangci-lint v2.8.0** - Static code analysis and style checks
- **govulncheck** - Scans dependencies for known security vulnerabilities
- **go test -race** - Runs unit tests with data race detector enabled
- **swag init** - Validates and generates Swagger documentation

### Build Flags

- `CGO_ENABLED=0` - Static binary without libc dependencies
- `-ldflags="-s -w"` - Strip debug symbols (~30% smaller binary)

## Systemd Service Configuration

Tested on Ubuntu 24.04 LTS (Noble Numbat) with systemd.

Create `/etc/systemd/system/<your-service-name>.service`:

```ini
[Unit]
Description=<Your Service Description>
After=network.target postgresql.service

[Service]
Type=simple
User=<user>
WorkingDirectory=<path/to/your/app>
ExecStart=<path/to/your/app/binary-name>
Restart=always
RestartSec=5
Environment=GIN_MODE=release
EnvironmentFile=-<path/to/your/app/.env>
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### Enable and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable <your-service-name>.service
sudo systemctl start <your-service-name>.service
sudo systemctl status <your-service-name>.service
journalctl -u <your-service-name>.service -f
```