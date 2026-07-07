package app

import (
	"strings"
	"testing"
)

func TestCmdVenueRequiresLat(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdVenue([]string{"--to", "123", "--long", "-73.98", "--title", "MSG", "--address", "4 Penn Plaza"})
		if exitCode != 2 {
			t.Fatalf("CmdVenue exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--lat is required") {
		t.Fatalf("CmdVenue output = %q, want --lat required message", output)
	}
}

func TestCmdVenueRequiresLong(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdVenue([]string{"--to", "123", "--lat", "40.75", "--title", "MSG", "--address", "4 Penn Plaza"})
		if exitCode != 2 {
			t.Fatalf("CmdVenue exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--long is required") {
		t.Fatalf("CmdVenue output = %q, want --long required message", output)
	}
}

func TestCmdVenueRequiresTitle(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdVenue([]string{"--to", "123", "--lat", "40.75", "--long", "-73.98", "--address", "4 Penn Plaza"})
		if exitCode != 2 {
			t.Fatalf("CmdVenue exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--title is required") {
		t.Fatalf("CmdVenue output = %q, want --title required message", output)
	}
}

func TestCmdVenueRequiresAddress(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdVenue([]string{"--to", "123", "--lat", "40.75", "--long", "-73.98", "--title", "MSG"})
		if exitCode != 2 {
			t.Fatalf("CmdVenue exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--address is required") {
		t.Fatalf("CmdVenue output = %q, want --address required message", output)
	}
}

func TestCmdVenueRejectsInvalidLat(t *testing.T) {
	setTestConfigHome(t)

	exitCode := CmdVenue([]string{"--to", "123", "--lat", "north", "--long", "-73.98", "--title", "MSG", "--address", "4 Penn Plaza"})
	if exitCode != 2 {
		t.Fatalf("CmdVenue exitCode = %d, want 2", exitCode)
	}
}
