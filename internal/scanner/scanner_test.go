package scanner

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestScanFindsOpenLocalPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	result, err := Scan(context.Background(), "127.0.0.1", Options{
		Ports:       []int{port},
		Concurrency: 4,
		Timeout:     time.Second,
		MaxHosts:    4,
	}, nil)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.OpenHostCount != 1 {
		t.Fatalf("OpenHostCount = %d, want 1", result.OpenHostCount)
	}
	if len(result.Hosts) != 1 {
		t.Fatalf("len(Hosts) = %d, want 1", len(result.Hosts))
	}
	if got := result.Hosts[0].OpenPorts[0].Port; got != port {
		t.Fatalf("open port = %d, want %d", got, port)
	}
}

func TestProbeTCPClosedPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if ProbeTCP(context.Background(), "127.0.0.1", port, 100*time.Millisecond) {
		t.Fatalf("ProbeTCP unexpectedly connected to closed port %s", strconv.Itoa(port))
	}
}
