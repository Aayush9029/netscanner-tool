package config

import "time"

const (
	DefaultConcurrency   = 512
	DiscoveryConcurrency = 256
	MaxDiscoveryHosts    = 2048
	MaxScanHosts         = 4096
)

const (
	DefaultTimeout   = 350 * time.Millisecond
	DiscoveryTimeout = 180 * time.Millisecond
)

func DefaultPorts() []int {
	return []int{
		20, 21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 389, 443,
		445, 465, 515, 548, 554, 587, 631, 993, 995, 1433, 1521, 2049,
		2375, 3000, 3306, 3389, 5000, 5432, 5900, 5985, 6379, 7000,
		7001, 8000, 8080, 8443, 8888, 9000, 9090, 9100, 9200, 9300,
		10000, 11211, 27017, 32400, 50070, 62078,
	}
}

func ProbePorts() []int {
	return []int{22, 53, 80, 443, 445, 548, 631, 3000, 5000, 8080, 8443, 62078}
}
