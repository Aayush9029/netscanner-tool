package scanner

import (
	"reflect"
	"testing"
)

func TestParsePortsDedupesSortsAndExpandsRanges(t *testing.T) {
	got, err := ParsePorts("443,22,80,8000-8002,22")
	if err != nil {
		t.Fatalf("ParsePorts returned error: %v", err)
	}

	want := []int{22, 80, 443, 8000, 8001, 8002}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ports = %#v, want %#v", got, want)
	}
}

func TestParsePortsRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"", "0", "65536", "80-", "90-80", "abc"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParsePorts(raw); err == nil {
				t.Fatalf("ParsePorts(%q) returned nil error", raw)
			}
		})
	}
}
