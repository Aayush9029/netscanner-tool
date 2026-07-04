package discovery

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Aayush9029/netscanner-tool/internal/scanner"
)

type Options struct {
	ProbePorts    []int
	Concurrency   int
	Timeout       time.Duration
	MaxProbeHosts int
}

type Network struct {
	InterfaceName string
	IP            net.IP
	CIDR          *net.IPNet
	Gateway       net.IP
}

type Suggestion struct {
	Label  string
	Target string
	Detail string
	Kind   string
}

type passiveHost struct {
	IP       string
	MAC      string
	Hostname string
}

func SmartSuggestions(opts Options) ([]Suggestion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	network, err := ActiveNetwork(ctx)
	if err != nil {
		return nil, err
	}

	suggestions := []Suggestion{{
		Label:  fmt.Sprintf("Current subnet %s", network.CIDR.String()),
		Target: network.CIDR.String(),
		Detail: fmt.Sprintf("%s on %s", network.IP.String(), network.InterfaceName),
		Kind:   "subnet",
	}}
	if network.Gateway != nil {
		suggestions = append(suggestions, Suggestion{
			Label:  "Gateway",
			Target: network.Gateway.String(),
			Detail: "default route",
			Kind:   "host",
		})
	}

	passive := PassiveHosts(ctx, network)
	for _, host := range passive {
		suggestions = appendHostSuggestion(suggestions, host.IP, passiveDetail(host), "passive")
	}

	probed := ProbeNetwork(ctx, network, opts)
	for _, host := range probed {
		suggestions = appendHostSuggestion(suggestions, host, "responded to TCP probe", "probe")
	}

	suggestions = append(suggestions, Suggestion{
		Label:  "Manual target",
		Target: "",
		Detail: "type an IP, hostname, CIDR, or comma-separated list",
		Kind:   "manual",
	})

	return dedupeSuggestions(suggestions), nil
}

func FallbackSuggestions() []Suggestion {
	return []Suggestion{
		{Label: "Localhost", Target: "127.0.0.1", Detail: "this Mac", Kind: "host"},
		{Label: "Manual target", Target: "", Detail: "type an IP, hostname, CIDR, or comma-separated list", Kind: "manual"},
	}
}

func ActiveNetwork(ctx context.Context) (Network, error) {
	defaultInterface, gateway := defaultRoute(ctx)
	if defaultInterface != "" {
		if network, err := networkForInterface(defaultInterface, gateway); err == nil {
			return network, nil
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return Network{}, err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		network, err := networkForInterface(iface.Name, nil)
		if err == nil {
			return network, nil
		}
	}
	return Network{}, fmt.Errorf("no active IPv4 network found")
}

func defaultRoute(ctx context.Context) (string, net.IP) {
	cmd := exec.CommandContext(ctx, "route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	var iface string
	var gateway net.IP
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "interface:") {
			iface = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
		}
		if strings.HasPrefix(line, "gateway:") {
			gateway = net.ParseIP(strings.TrimSpace(strings.TrimPrefix(line, "gateway:")))
		}
	}
	return iface, gateway
}

func networkForInterface(name string, gateway net.IP) (Network, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return Network{}, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return Network{}, err
	}
	for _, addr := range addrs {
		ip, cidr, ok := parseIPv4Addr(addr)
		if !ok {
			continue
		}
		cidr.IP = ip.Mask(cidr.Mask)
		return Network{
			InterfaceName: name,
			IP:            ip,
			CIDR:          cidr,
			Gateway:       gateway,
		}, nil
	}
	return Network{}, fmt.Errorf("interface %s has no IPv4 address", name)
}

func parseIPv4Addr(addr net.Addr) (net.IP, *net.IPNet, bool) {
	ipNet, ok := addr.(*net.IPNet)
	if !ok {
		return nil, nil, false
	}
	ip := ipNet.IP.To4()
	if ip == nil {
		return nil, nil, false
	}
	return ip, &net.IPNet{IP: ip, Mask: ipNet.Mask}, true
}

func PassiveHosts(ctx context.Context, network Network) []passiveHost {
	cmd := exec.CommandContext(ctx, "arp", "-an")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lineRE := regexp.MustCompile(`^(\S+)\s+\((\d+\.\d+\.\d+\.\d+)\)\s+at\s+([0-9a-fA-F:]{17})`)
	var hosts []passiveHost
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(strings.ToLower(line), "incomplete") {
			continue
		}
		match := lineRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) != 4 {
			continue
		}
		ip := net.ParseIP(match[2])
		if ip == nil || !network.CIDR.Contains(ip) {
			continue
		}
		hostname := ""
		if match[1] != "?" {
			hostname = match[1]
		}
		hosts = append(hosts, passiveHost{
			IP:       match[2],
			MAC:      strings.ToLower(match[3]),
			Hostname: hostname,
		})
	}
	return hosts
}

func ProbeNetwork(ctx context.Context, network Network, opts Options) []string {
	if opts.Concurrency < 1 {
		opts.Concurrency = 128
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 180 * time.Millisecond
	}
	if opts.MaxProbeHosts <= 0 {
		opts.MaxProbeHosts = 2048
	}
	if len(opts.ProbePorts) == 0 {
		opts.ProbePorts = []int{22, 80, 443}
	}

	hosts, err := scanner.ExpandTargets(network.CIDR.String(), opts.MaxProbeHosts)
	if err != nil {
		return nil
	}

	jobs := make(chan string)
	results := make(chan string, len(hosts))
	var workers sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for host := range jobs {
				for _, port := range opts.ProbePorts {
					if scanner.ProbeTCP(ctx, host, port, opts.Timeout) {
						results <- host
						break
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, host := range hosts {
			select {
			case <-ctx.Done():
				return
			case jobs <- host:
			}
		}
	}()

	workers.Wait()
	close(results)

	seen := map[string]bool{}
	var alive []string
	for host := range results {
		if !seen[host] {
			seen[host] = true
			alive = append(alive, host)
		}
	}
	sort.Slice(alive, func(i, j int) bool {
		return lessIP(alive[i], alive[j])
	})
	return alive
}

func appendHostSuggestion(suggestions []Suggestion, target string, detail string, kind string) []Suggestion {
	if target == "" {
		return suggestions
	}
	return append(suggestions, Suggestion{
		Label:  target,
		Target: target,
		Detail: detail,
		Kind:   kind,
	})
}

func passiveDetail(host passiveHost) string {
	var parts []string
	if host.Hostname != "" {
		parts = append(parts, host.Hostname)
	}
	if host.MAC != "" {
		parts = append(parts, host.MAC)
	}
	if len(parts) == 0 {
		return "seen in neighbor table"
	}
	return strings.Join(parts, "  ")
}

func dedupeSuggestions(input []Suggestion) []Suggestion {
	seen := map[string]bool{}
	var output []Suggestion
	for _, suggestion := range input {
		key := "host:" + suggestion.Target
		if suggestion.Kind == "manual" {
			key = "manual"
		}
		if suggestion.Kind == "subnet" {
			key = "subnet:" + suggestion.Target
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		output = append(output, suggestion)
	}
	return output
}

func lessIP(a, b string) bool {
	ipA := net.ParseIP(a).To4()
	ipB := net.ParseIP(b).To4()
	if ipA == nil || ipB == nil {
		return a < b
	}
	for i := 0; i < net.IPv4len; i++ {
		if ipA[i] == ipB[i] {
			continue
		}
		return ipA[i] < ipB[i]
	}
	return false
}
