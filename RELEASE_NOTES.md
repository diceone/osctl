## 🔒 Security Enhancements

- **Environment Variable Configuration**: Configure credentials and port via `OSCTL_USERNAME`, `OSCTL_PASSWORD`, and `OSCTL_PORT`
- **Proper Basic Auth**: Added WWW-Authenticate headers for standard-compliant HTTP authentication
- **Input Validation**: Service names are validated to prevent command injection attacks
- **Parameter Limits**: Length limits enforced on API parameters

## 🛠️ Stability Improvements

- **Error Handling**: Replaced `log.Fatalf()` with error returns - API server no longer crashes on individual command failures
- **Graceful Degradation**: Server continues running even when individual system commands fail
- **Better Error Messages**: Improved error reporting for easier troubleshooting

## ⚙️ Configuration

All configuration now via environment variables:
- `OSCTL_PORT`: Server port (default: 12000)
- `OSCTL_USERNAME`: Basic auth username (default: admin)
- `OSCTL_PASSWORD`: Basic auth password (default: password)

Example:
```bash
export OSCTL_PORT=8080
export OSCTL_USERNAME=myuser
export OSCTL_PASSWORD=securepass
./osctl api
```

## 🐧 OS Compatibility

- **Modern OS Detection**: Uses `/etc/os-release` (modern standard) with fallback to legacy files
- **Better Distribution Support**: Improved detection for RHEL, CentOS, Fedora, Ubuntu, Debian, SUSE, openSUSE

## 📖 Documentation

- **AI Agent Instructions**: Added `.github/copilot-instructions.md` for AI coding assistants
- **Security Best Practices**: Comprehensive security section in README
- **Deployment Examples**: Updated examples with environment variable configuration

## 🔄 Breaking Changes

**None** - This release is fully backward compatible. Default values match previous behavior.

## 📦 Installation

```bash
# Download binary
wget https://github.com/diceone/osctl/releases/download/v0.0.6/osctl

# Or build from source
git clone https://github.com/diceone/osctl.git
cd osctl
git checkout v0.0.6
go build -o osctl main.go auth.go metrics.go handlers.go system_info.go services.go
```

## 🚀 What's Changed

Full Changelog: https://github.com/diceone/osctl/compare/v0.0.5...v0.0.6
