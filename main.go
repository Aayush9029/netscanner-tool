package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Aayush9029/netscanner-tool/internal/config"
	"github.com/Aayush9029/netscanner-tool/internal/discovery"
	"github.com/Aayush9029/netscanner-tool/internal/scanner"
	"github.com/Aayush9029/netscanner-tool/internal/tui"
	"github.com/Aayush9029/netscanner-tool/internal/ui"
)

var version = "dev"

type cliOptions struct {
	ports       []int
	concurrency int
	timeout     time.Duration
	json        bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		ui.Fatalf("%s", err)
	}
}

func run(args []string) error {
	opts := defaultCLIOptions()

	if len(args) == 0 {
		return cmdInteractive(opts)
	}

	switch args[0] {
	case "-h", "--help", "help":
		showHelp()
		return nil
	case "-v", "--version", "version":
		fmt.Printf("netscanner %s\n", version)
		return nil
	case "scan":
		opts, targets, err := parseOptions(args[1:], opts)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return errors.New("scan requires a target")
		}
		return cmdScan(strings.Join(targets, ","), opts)
	default:
		opts, targets, err := parseOptions(args, opts)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return cmdInteractive(opts)
		}
		return cmdScan(strings.Join(targets, ","), opts)
	}
}

func defaultCLIOptions() cliOptions {
	return cliOptions{
		ports:       config.DefaultPorts(),
		concurrency: config.DefaultConcurrency,
		timeout:     config.DefaultTimeout,
	}
}

func parseOptions(args []string, opts cliOptions) (cliOptions, []string, error) {
	var targets []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-p", "--ports":
			i++
			if i >= len(args) {
				return opts, nil, errors.New("--ports requires a value")
			}
			ports, err := scanner.ParsePorts(args[i])
			if err != nil {
				return opts, nil, err
			}
			opts.ports = ports
		case "-c", "--concurrency":
			i++
			if i >= len(args) {
				return opts, nil, errors.New("--concurrency requires a value")
			}
			value, err := strconv.Atoi(args[i])
			if err != nil || value < 1 {
				return opts, nil, fmt.Errorf("invalid concurrency: %s", args[i])
			}
			opts.concurrency = value
		case "--timeout":
			i++
			if i >= len(args) {
				return opts, nil, errors.New("--timeout requires a value")
			}
			timeout, err := parseTimeout(args[i])
			if err != nil {
				return opts, nil, err
			}
			opts.timeout = timeout
		case "--json":
			opts.json = true
		case "-h", "--help", "help":
			showHelp()
			os.Exit(0)
		case "-v", "--version", "version":
			fmt.Printf("netscanner %s\n", version)
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				return opts, nil, fmt.Errorf("unknown option: %s", arg)
			}
			targets = append(targets, arg)
		}
	}

	return opts, targets, nil
}

func parseTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, errors.New("timeout cannot be empty")
	}
	if duration, err := time.ParseDuration(raw); err == nil {
		return duration, nil
	}
	seconds, err := strconv.ParseFloat(raw, 64)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("invalid timeout: %s", raw)
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

func cmdInteractive(opts cliOptions) error {
	ui.Header("netscanner")
	ui.Dimf("looking up this network")
	fmt.Println()

	suggestions, err := discovery.SmartSuggestions(discovery.Options{
		ProbePorts:    config.ProbePorts(),
		Concurrency:   config.DiscoveryConcurrency,
		Timeout:       config.DiscoveryTimeout,
		MaxProbeHosts: config.MaxDiscoveryHosts,
	})
	if err != nil {
		ui.Warn(err.Error())
	}
	if len(suggestions) == 0 {
		suggestions = discovery.FallbackSuggestions()
	}

	return tui.Run(suggestions, scanner.Options{
		Ports:       opts.ports,
		Concurrency: opts.concurrency,
		Timeout:     opts.timeout,
		MaxHosts:    config.MaxScanHosts,
	})
}

func cmdScan(target string, opts cliOptions) error {
	scanOpts := scanner.Options{
		Ports:       opts.ports,
		Concurrency: opts.concurrency,
		Timeout:     opts.timeout,
		MaxHosts:    config.MaxScanHosts,
	}

	if opts.json {
		result, err := scanner.Scan(context.Background(), target, scanOpts, nil)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	return tui.RunScan(target, scanOpts)
}

func showHelp() {
	fmt.Println()
	ui.Header("netscanner")
	ui.Dimf("smart local network scanner")
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Println("    netscanner")
	fmt.Println("    netscanner scan <target> [options]")
	fmt.Println("    netscanner <target> [options]")
	fmt.Println()
	fmt.Println("TARGETS")
	fmt.Println("    192.168.1.10")
	fmt.Println("    192.168.1.0/24")
	fmt.Println("    router.local")
	fmt.Println("    192.168.1.1,192.168.1.20")
	fmt.Println()
	fmt.Println("OPTIONS")
	fmt.Println("    -p, --ports LIST       Ports or ranges, like 22,80,443,8000-8010")
	fmt.Println("    -c, --concurrency N    Parallel TCP checks (default: 512)")
	fmt.Println("    --timeout DURATION     TCP timeout, like 300ms or 1s (default: 350ms)")
	fmt.Println("    --json                 Print JSON instead of the live UI")
	fmt.Println("    -h, --help             Show this help")
	fmt.Println("    -v, --version          Show version")
	fmt.Println()
}
