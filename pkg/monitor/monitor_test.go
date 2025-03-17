package monitor

import (
	"testing"
	"time"
)

func TestNewMonitor(t *testing.T) {
	mon := NewMonitor(5 * time.Second)
	if mon == nil {
		t.Fatal("NewMonitor returned nil")
	}

	// 기본 속성 확인
	if mon.interval != 5*time.Second {
		t.Errorf("Expected interval to be 5s, got %v", mon.interval)
	}

	if len(mon.devices) != 0 {
		t.Errorf("Expected empty devices list, got %v", mon.devices)
	}

	if mon.running {
		t.Error("Expected monitor to be not running initially")
	}
}

func TestAddDevice(t *testing.T) {
	mon := NewMonitor(5 * time.Second)
	mon.AddDevice("C:\\")

	if len(mon.devices) != 1 {
		t.Fatalf("Expected 1 device, got %d", len(mon.devices))
	}

	if mon.devices[0] != "C:\\" {
		t.Errorf("Expected device to be C:\\, got %s", mon.devices[0])
	}
}

func TestSetFileFilters(t *testing.T) {
	mon := NewMonitor(5 * time.Second)
	filters := []string{".exe", ".dll", ".sys"}
	mon.SetFileFilters(filters)

	if len(mon.fileFilters) != len(filters) {
		t.Fatalf("Expected %d filters, got %d", len(filters), len(mon.fileFilters))
	}

	for i, filter := range filters {
		if mon.fileFilters[i] != filter {
			t.Errorf("Expected filter[%d] to be %s, got %s", i, filter, mon.fileFilters[i])
		}
	}
}
