Name:           osctl
Version:        1.0.0
Release:        1%{?dist}
Summary:        osctl system administration tool

License:        MIT
URL:            https://github.com/yourusername/osctl
Source0:        %{name}-%{version}.tar.gz

Requires:       systemd

%description
osctl is a command-line tool for Linux system administration. It provides various commands to monitor and manage the system, including checking RAM and disk usage, managing services, viewing top processes, checking system logs, and more.

%prep
%setup -q

%build

%install
# Create necessary directories
install -d %{buildroot}/usr/local/bin
install -d %{buildroot}/etc/systemd/system

# Install the binary
install -m 0755 osctl %{buildroot}/usr/local/bin/osctl

# Install the systemd unit file
install -m 0644 osctl.service %{buildroot}/etc/systemd/system/osctl.service

%files
%license LICENSE
%doc README.md
/usr/local/bin/osctl
/etc/systemd/system/osctl.service

%post
# Reload systemd to recognize the new service
systemctl daemon-reload

%preun
if [ $1 -eq 0 ]; then
    # Stop the service if it is running
    systemctl stop osctl.service
    # Disable the service
    systemctl disable osctl.service
fi

%postun
if [ $1 -eq 0 ]; then
    # Remove the systemd unit file and reload systemd
    systemctl daemon-reload
fi

%changelog
* Tue May 21 2024 Your Name <youremail@example.com> - 1.0.0-1
- Initial package
