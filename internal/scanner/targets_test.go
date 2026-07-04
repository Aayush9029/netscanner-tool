package scanner

import (
	"reflect"
	"testing"
)

func TestExpandTargetsCIDR(t *testing.T) {
	got, err := ExpandTargets("192.168.1.0/30", 16)
	if err != nil {
		t.Fatalf("ExpandTargets returned error: %v", err)
	}

	want := []string{"192.168.1.1", "192.168.1.2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hosts = %#v, want %#v", got, want)
	}
}

func TestExpandTargetsDedupesAndSorts(t *testing.T) {
	got, err := ExpandTargets("192.168.1.20,192.168.1.2,192.168.1.20,example.local", 16)
	if err != nil {
		t.Fatalf("ExpandTargets returned error: %v", err)
	}

	want := []string{"192.168.1.2", "192.168.1.20", "example.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hosts = %#v, want %#v", got, want)
	}
}

func TestExpandTargetsRejectsTooLargeCIDR(t *testing.T) {
	if _, err := ExpandTargets("10.0.0.0/16", 256); err == nil {
		t.Fatal("expected max-hosts error")
	}
}
