package scanner

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

func ExpandTargets(raw string, maxHosts int) ([]string, error) {
	var hosts []string
	seen := map[string]bool{}

	for _, part := range strings.Split(raw, ",") {
		target := strings.TrimSpace(part)
		if target == "" {
			continue
		}

		expanded, err := expandTarget(target, maxHosts)
		if err != nil {
			return nil, err
		}
		for _, host := range expanded {
			if !seen[host] {
				seen[host] = true
				hosts = append(hosts, host)
			}
		}
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("target cannot be empty")
	}
	sort.SliceStable(hosts, func(i, j int) bool {
		return lessHost(hosts[i], hosts[j])
	})
	return hosts, nil
}

func expandTarget(target string, maxHosts int) ([]string, error) {
	if strings.Contains(target, "/") {
		ip, network, err := net.ParseCIDR(target)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR target %q: %w", target, err)
		}
		if ip.To4() == nil {
			return nil, fmt.Errorf("IPv6 CIDR scanning is not supported yet: %s", target)
		}
		return expandIPv4CIDR(network, maxHosts)
	}

	if ip := net.ParseIP(target); ip != nil {
		if ip.To4() == nil {
			return nil, fmt.Errorf("IPv6 scanning is not supported yet: %s", target)
		}
		return []string{ip.String()}, nil
	}

	return []string{target}, nil
}

func expandIPv4CIDR(network *net.IPNet, maxHosts int) ([]string, error) {
	ip := network.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("only IPv4 CIDRs are supported")
	}

	ones, bits := network.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("only IPv4 CIDRs are supported")
	}

	total := 1 << uint(32-ones)
	if total > maxHosts+2 {
		return nil, fmt.Errorf("%s has %d addresses; max is %d", network.String(), total, maxHosts)
	}

	hosts := make([]string, 0, total)
	for current := cloneIP4(ip.Mask(network.Mask)); network.Contains(current); incrementIP(current) {
		if total > 2 && (isNetworkAddress(current, network) || isBroadcastAddress(current, network)) {
			continue
		}
		hosts = append(hosts, current.String())
	}
	return hosts, nil
}

func cloneIP4(ip net.IP) net.IP {
	out := make(net.IP, net.IPv4len)
	copy(out, ip.To4())
	return out
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			return
		}
	}
}

func isNetworkAddress(ip net.IP, network *net.IPNet) bool {
	return ip.Equal(network.IP.Mask(network.Mask))
}

func isBroadcastAddress(ip net.IP, network *net.IPNet) bool {
	broadcast := cloneIP4(network.IP)
	for i := range broadcast {
		broadcast[i] |= ^network.Mask[i]
	}
	return ip.Equal(broadcast)
}

func lessHost(a, b string) bool {
	ipA := net.ParseIP(a)
	ipB := net.ParseIP(b)
	if ipA4, ipB4 := ipA.To4(), ipB.To4(); ipA4 != nil && ipB4 != nil {
		for i := 0; i < net.IPv4len; i++ {
			if ipA4[i] == ipB4[i] {
				continue
			}
			return ipA4[i] < ipB4[i]
		}
		return false
	}
	return a < b
}
