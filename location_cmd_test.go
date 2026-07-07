package main

import (
	"strings"
	"testing"
)

func TestCmdLocationRequiresLat(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdLocation([]string{"--to", "123", "--long", "-73.9855"})
		if exitCode != 2 {
			t.Fatalf("cmdLocation exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--lat is required") {
		t.Fatalf("cmdLocation output = %q, want --lat required message", output)
	}
}

func TestCmdLocationRequiresLong(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdLocation([]string{"--to", "123", "--lat", "40.7580"})
		if exitCode != 2 {
			t.Fatalf("cmdLocation exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--long is required") {
		t.Fatalf("cmdLocation output = %q, want --long required message", output)
	}
}

func TestCmdLocationRejectsInvalidLat(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdLocation([]string{"--to", "123", "--lat", "north", "--long", "-73.9855"})
	if exitCode != 2 {
		t.Fatalf("cmdLocation exitCode = %d, want 2", exitCode)
	}
}

func TestCmdLocationRejectsInvalidLivePeriod(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdLocation([]string{"--to", "123", "--lat", "40.7580", "--long", "-73.9855", "--live-period", "soon"})
	if exitCode != 2 {
		t.Fatalf("cmdLocation exitCode = %d, want 2", exitCode)
	}
}
