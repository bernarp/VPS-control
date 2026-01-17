# VPS Control API

VPS management API built with Go and Gin framework.

## CI/CD Pipeline

GitHub Actions workflow with 4 stages:

1. Lint - golangci-lint + govulncheck + swagger generation
2. Test - Unit tests with race detector
3. Build - Cross-compile for Linux AMD64
4. Deploy - SCP binary to VPS + systemctl restart

### Required GitHub Secrets

- HOST - VPS hostname
- USERNAME - SSH user
- SSH_PRIVATE_KEY - Private key for authentication
- DEPLOY_PATH - Deployment directory (e.g. /opt/apps/vps-control)
- SERVICE_NAME - systemd service name (e.g. vps-api.service)

### Tests and Checks

- golangci-lint v2.8.0 - Static code analysis and style checks
- govulncheck - Scans dependencies for known security vulnerabilities
- go test -race - Runs unit tests with data race detector enabled
- swag init - Validates and generates Swagger documentation

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

```
sudo systemctl daemon-reload
sudo systemctl enable <your-service-name>.service
sudo systemctl start <your-service-name>.service
sudo systemctl status <your-service-name>.service
journalctl -u <your-service-name>.service -f
```
