package app

import (
	"strings"
	"testing"
)

func TestCmdMediaGroupRejectsTooFewFiles(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdMediaGroup([]string{"--to", "123", "--file", "AgACAgIAAx0FakeFileID"})
		if exitCode != 2 {
			t.Fatalf("CmdMediaGroup exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--file must be given 2-10 times (got 1)") {
		t.Fatalf("CmdMediaGroup output = %q, want too few files message", output)
	}
}

func TestCmdMediaGroupRejectsTooManyFiles(t *testing.T) {
	setTestConfigHome(t)

	args := []string{"--to", "123"}
	for _, file := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"} {
		args = append(args, "--file", file)
	}

	output := captureStdout(t, func() int {
		exitCode := CmdMediaGroup(args)
		if exitCode != 2 {
			t.Fatalf("CmdMediaGroup exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--file must be given 2-10 times (got 11)") {
		t.Fatalf("CmdMediaGroup output = %q, want too many files message", output)
	}
}

func TestCmdMediaGroupRejectsInvalidType(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdMediaGroup([]string{"--to", "123", "--type", "gif", "--file", "A", "--file", "B"})
		if exitCode != 2 {
			t.Fatalf("CmdMediaGroup exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, `invalid --type "gif"`) {
		t.Fatalf("CmdMediaGroup output = %q, want invalid type message", output)
	}
	if !strings.Contains(output, "valid values: photo, video, audio, document") {
		t.Fatalf("CmdMediaGroup output = %q, want valid values hint", output)
	}
}
