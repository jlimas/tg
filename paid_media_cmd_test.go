package main

import (
	"strings"
	"testing"
)

func TestCmdPaidMediaRequiresStarCount(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPaidMedia([]string{"--to", "123", "--file", "AgACAgIAAx0FakeFileID"})
		if exitCode != 2 {
			t.Fatalf("cmdPaidMedia exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--star-count is required") {
		t.Fatalf("cmdPaidMedia output = %q, want --star-count required message", output)
	}
}

func TestCmdPaidMediaRejectsNonPositiveStarCount(t *testing.T) {
	tests := []string{"0", "-5"}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			setTestConfigHome(t)

			output := captureStdout(t, func() int {
				exitCode := cmdPaidMedia([]string{"--to", "123", "--star-count", value, "--file", "AgACAgIAAx0FakeFileID"})
				if exitCode != 2 {
					t.Fatalf("cmdPaidMedia exitCode = %d, want 2", exitCode)
				}
				return exitCode
			})

			if !strings.Contains(output, "--star-count must be a positive integer") {
				t.Fatalf("cmdPaidMedia output = %q, want positive integer message", output)
			}
		})
	}
}

func TestCmdPaidMediaRejectsInvalidType(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPaidMedia([]string{"--to", "123", "--star-count", "50", "--type", "audio", "--file", "AgACAgIAAx0FakeFileID"})
		if exitCode != 2 {
			t.Fatalf("cmdPaidMedia exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, `invalid --type "audio"`) {
		t.Fatalf("cmdPaidMedia output = %q, want invalid type message", output)
	}
	if !strings.Contains(output, "valid values: photo, video") {
		t.Fatalf("cmdPaidMedia output = %q, want valid values hint", output)
	}
}

func TestCmdPaidMediaRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPaidMedia([]string{"--to", "123", "--star-count", "50"})
		if exitCode != 2 {
			t.Fatalf("cmdPaidMedia exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--file must be given at least once") {
		t.Fatalf("cmdPaidMedia output = %q, want --file required message", output)
	}
}
