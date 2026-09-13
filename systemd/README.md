# Running osctl as a systemd Service

osctl ships a systemd unit file in [`osctl.service`](osctl.service) and the
deb/rpm packages published with each release install it automatically.

## Option 1: Install a package (recommended)

```bash
# Debian/Ubuntu
sudo apt install ./osctl_<version>_amd64.deb

# RHEL/CentOS/Fedora/SUSE
sudo rpm -i osctl_<version>_amd64.rpm
```

The package installs the binary to `/usr/bin/osctl` and the unit file to
`/lib/systemd/system/osctl.service`. Configure credentials via
[environment variables](../README.md#configuration) or a config file
(`OSCTL_CONFIG`), then:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now osctl
```

## Option 2: Manual setup

1. **Build the binary** (all Go files live in one package):

   ```bash
   git clone https://github.com/diceone/osctl.git
   cd osctl
   go build -o osctl .
   sudo mv osctl /usr/local/bin/osctl
   ```

2. **Install the unit file** from this directory:

   ```bash
   sudo cp systemd/osctl.service /etc/systemd/system/osctl.service
   sudo systemctl daemon-reload
   ```

   The unit file starts `/usr/bin/osctl api` as root on port 12000 (the same
   path the deb/rpm packages use). Adjust `ExecStart` if you installed the
   binary elsewhere.

3. **Configure credentials** — the server reads `OSCTL_PORT`, `OSCTL_USERNAME`,
   `OSCTL_PASSWORD`, `OSCTL_API_TOKEN` and other variables. Add an
   `EnvironmentFile=` line to the `[Service]` section, e.g.:

   ```ini
   EnvironmentFile=/etc/osctl/osctl.env
   ```

4. **Enable and start**:

   ```bash
   sudo systemctl enable --now osctl
   ```

5. **Verify**:

   ```bash
   sudo systemctl status osctl
   curl -u admin:password http://localhost:12000/health
   ```

## Logging

The unit directs stdout/stderr to the system log, so osctl API logs are
viewable via journalctl:

```bash
sudo journalctl -u osctl -f
```
