package app

import (
	"strings"
	"testing"
)

func TestCmdContactRequiresPhone(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdContact([]string{"--to", "123", "--first-name", "Ada"})
		if exitCode != 2 {
			t.Fatalf("CmdContact exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--phone is required") {
		t.Fatalf("CmdContact output = %q, want --phone required message", output)
	}
}

func TestCmdContactRequiresFirstName(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdContact([]string{"--to", "123", "--phone", "+15551234567"})
		if exitCode != 2 {
			t.Fatalf("CmdContact exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--first-name is required") {
		t.Fatalf("CmdContact output = %q, want --first-name required message", output)
	}
}
