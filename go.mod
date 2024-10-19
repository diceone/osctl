module osctl

go 1.20

require (
	github.com/prometheus/client_golang v1.19.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/vishvananda/netlink v1.1.0
)

replace (
	github.com/beorn7/perks => github.com/beorn7/perks@v1.0.1 // indirect
	github.com/cespare/xxhash/v2 => github.com/cespare/xxhash/v2@v2.2.0 // indirect
)
