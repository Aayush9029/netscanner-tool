package scanner

import (
	"context"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Options struct {
	Ports       []int
	Concurrency int
	Timeout     time.Duration
	MaxHosts    int
}

type EventKind string

const (
	EventStarted EventKind = "started"
	EventPort    EventKind = "port"
	EventOpen    EventKind = "open"
)

type Event struct {
	Kind            EventKind
	Target          string
	Host            string
	Port            int
	Service         string
	Open            bool
	CompletedChecks int
	TotalChecks     int
	Message         string
}

type PortResult struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
}

type HostResult struct {
	Host       string       `json:"host"`
	Hostname   string       `json:"hostname,omitempty"`
	OpenPorts  []PortResult `json:"open_ports"`
	DurationMS int64        `json:"duration_ms"`
}

type Result struct {
	Target        string       `json:"target"`
	StartedAt     time.Time    `json:"started_at"`
	FinishedAt    time.Time    `json:"finished_at"`
	DurationMS    int64        `json:"duration_ms"`
	HostsScanned  int          `json:"hosts_scanned"`
	PortsScanned  int          `json:"ports_scanned"`
	OpenHostCount int          `json:"open_host_count"`
	OpenPortCount int          `json:"open_port_count"`
	Concurrency   int          `json:"concurrency"`
	TimeoutMS     int64        `json:"timeout_ms"`
	Ports         []int        `json:"ports"`
	Hosts         []HostResult `json:"hosts"`
}

type job struct {
	host string
	port int
}

type hostState struct {
	started time.Time
	ports   []PortResult
}

func Scan(ctx context.Context, target string, opts Options, events chan<- Event) (Result, error) {
	if opts.Concurrency < 1 {
		opts.Concurrency = 128
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 350 * time.Millisecond
	}
	if opts.MaxHosts <= 0 {
		opts.MaxHosts = 4096
	}
	if len(opts.Ports) == 0 {
		opts.Ports = []int{22, 80, 443}
	}

	hosts, err := ExpandTargets(target, opts.MaxHosts)
	if err != nil {
		return Result{}, err
	}

	started := time.Now()
	totalChecks := len(hosts) * len(opts.Ports)
	emit(events, Event{Kind: EventStarted, Target: target, TotalChecks: totalChecks})

	states := make(map[string]*hostState, len(hosts))
	for _, host := range hosts {
		states[host] = &hostState{started: time.Now()}
	}

	jobs := make(chan job)
	var mu sync.Mutex
	var completed int

	workerCount := opts.Concurrency
	if workerCount > totalChecks {
		workerCount = totalChecks
	}
	if workerCount < 1 {
		workerCount = 1
	}

	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for task := range jobs {
				open := ProbeTCP(ctx, task.host, task.port, opts.Timeout)
				service := ServiceName(task.port)

				mu.Lock()
				completed++
				checks := completed
				if open {
					states[task.host].ports = append(states[task.host].ports, PortResult{
						Port:    task.port,
						Service: service,
					})
				}
				mu.Unlock()

				kind := EventPort
				if open {
					kind = EventOpen
				}
				emit(events, Event{
					Kind:            kind,
					Target:          target,
					Host:            task.host,
					Port:            task.port,
					Service:         service,
					Open:            open,
					CompletedChecks: checks,
					TotalChecks:     totalChecks,
				})
			}
		}()
	}

sendJobs:
	for _, host := range hosts {
		for _, port := range opts.Ports {
			select {
			case <-ctx.Done():
				break sendJobs
			case jobs <- job{host: host, port: port}:
			}
		}
	}
	close(jobs)
	workers.Wait()

	finished := time.Now()
	result := Result{
		Target:        target,
		StartedAt:     started,
		FinishedAt:    finished,
		DurationMS:    finished.Sub(started).Milliseconds(),
		HostsScanned:  len(hosts),
		PortsScanned:  completed,
		Concurrency:   opts.Concurrency,
		TimeoutMS:     opts.Timeout.Milliseconds(),
		Ports:         append([]int(nil), opts.Ports...),
		OpenHostCount: 0,
		OpenPortCount: 0,
	}

	for _, host := range hosts {
		state := states[host]
		if len(state.ports) == 0 {
			continue
		}
		sort.Slice(state.ports, func(i, j int) bool {
			return state.ports[i].Port < state.ports[j].Port
		})
		hostname := lookupHostname(ctx, host)
		result.Hosts = append(result.Hosts, HostResult{
			Host:       host,
			Hostname:   hostname,
			OpenPorts:  state.ports,
			DurationMS: finished.Sub(state.started).Milliseconds(),
		})
		result.OpenHostCount++
		result.OpenPortCount += len(state.ports)
	}

	sort.Slice(result.Hosts, func(i, j int) bool {
		return lessHost(result.Hosts[i].Host, result.Hosts[j].Host)
	})

	if ctx.Err() != nil && completed < totalChecks {
		return result, ctx.Err()
	}
	return result, nil
}

func ProbeTCP(ctx context.Context, host string, port int, timeout time.Duration) bool {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(probeCtx, "tcp", net.JoinHostPort(host, intToString(port)))
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func lookupHostname(ctx context.Context, host string) string {
	lookupCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	names, err := net.DefaultResolver.LookupAddr(lookupCtx, host)
	if err != nil || len(names) == 0 {
		return ""
	}
	return trimTrailingDot(names[0])
}

func trimTrailingDot(value string) string {
	if len(value) > 0 && value[len(value)-1] == '.' {
		return value[:len(value)-1]
	}
	return value
}

func emit(events chan<- Event, event Event) {
	if events == nil {
		return
	}
	select {
	case events <- event:
	default:
	}
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
