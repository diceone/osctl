Certainly! Here's a `systemd` unit file to run the `osctl` API as a service.

### `osctl.service`

### Instructions to Set Up the `osctl` API Service

1. **Create the `osctl` binary**:
   Make sure you have built the `osctl` binary and placed it in `/root/osctl` (or adjust the paths accordingly).

   ```bash
   cd /path/to/osctl
   go build -o osctl osctl.go
   mv osctl /root/osctl
   ```

2. **Create the systemd service file**:
   Create a new file at `/etc/systemd/system/osctl.service` and copy the unit file content above into it.

   ```bash
   sudo nano /etc/systemd/system/osctl.service
   ```

   Paste the content:

   ```ini
   [Unit]
   Description=osctl API Service
   After=network.target

   [Service]
   Type=simple
   User=root
   WorkingDirectory=/usr/local/bin/
   ExecStart=/usr/local/bin/osctl api -username admin -password admin
   Restart=on-failure
   RestartSec=10
   StandardOutput=syslog
   StandardError=syslog
   SyslogIdentifier=osctl

   [Install]
   WantedBy=multi-user.target
   ```

3. **Reload systemd to recognize the new service**:
   ```bash
   sudo systemctl daemon-reload
   ```

4. **Enable the service to start on boot**:
   ```bash
   sudo systemctl enable osctl
   ```

5. **Start the service**:
   ```bash
   sudo systemctl start osctl
   ```

6. **Check the status of the service**:
   ```bash
   sudo systemctl status osctl
   ```

   You should see output indicating that the service is running.

### Logging

The `StandardOutput` and `StandardError` directives in the service file direct the output and error logs to the system log. You can view the logs using `journalctl`:

```bash
sudo journalctl -u osctl -f
```

This will display the logs for the `osctl` service in real-time.

By following these steps, you can set up `osctl` to run as a service using `systemd`, ensuring it starts automatically on boot and can be managed using standard `systemd` commands.
