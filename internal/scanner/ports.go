package scanner

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ParsePorts(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("ports cannot be empty")
	}

	seen := map[int]bool{}
	var ports []int
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty port in %q", raw)
		}
		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}
			start, err := parsePort(bounds[0])
			if err != nil {
				return nil, err
			}
			end, err := parsePort(bounds[1])
			if err != nil {
				return nil, err
			}
			if start > end {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}
			for port := start; port <= end; port++ {
				if !seen[port] {
					seen[port] = true
					ports = append(ports, port)
				}
			}
			continue
		}

		port, err := parsePort(part)
		if err != nil {
			return nil, err
		}
		if !seen[port] {
			seen[port] = true
			ports = append(ports, port)
		}
	}

	sort.Ints(ports)
	return ports, nil
}

func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port: %s", strings.TrimSpace(raw))
	}
	return port, nil
}

func ServiceName(port int) string {
	if name, ok := serviceNames[port]; ok {
		return name
	}
	return "tcp"
}

var serviceNames = map[int]string{
	20:    "ftp-data",
	21:    "ftp",
	22:    "ssh",
	23:    "telnet",
	25:    "smtp",
	53:    "dns",
	80:    "http",
	110:   "pop3",
	111:   "rpcbind",
	135:   "msrpc",
	139:   "netbios",
	143:   "imap",
	389:   "ldap",
	443:   "https",
	445:   "smb",
	465:   "smtps",
	515:   "printer",
	548:   "afp",
	554:   "rtsp",
	587:   "submission",
	631:   "ipp",
	993:   "imaps",
	995:   "pop3s",
	1433:  "mssql",
	1521:  "oracle",
	2049:  "nfs",
	2375:  "docker",
	3000:  "dev",
	3306:  "mysql",
	3389:  "rdp",
	5000:  "upnp/dev",
	5432:  "postgres",
	5900:  "vnc",
	5985:  "winrm",
	6379:  "redis",
	7000:  "dev",
	7001:  "dev",
	8000:  "http-alt",
	8080:  "http-alt",
	8443:  "https-alt",
	8888:  "notebook",
	9000:  "admin",
	9090:  "admin",
	9100:  "jetdirect",
	9200:  "elasticsearch",
	9300:  "elasticsearch",
	10000: "webmin",
	11211: "memcached",
	27017: "mongodb",
	32400: "plex",
	50070: "hdfs",
	62078: "iphone-sync",
}
