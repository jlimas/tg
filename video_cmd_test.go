package main

import "testing"

func TestCmdVideoRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVideo([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdVideo exitCode = %d, want 2", exitCode)
	}
}

func TestCmdVideoRejectsInvalidDuration(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVideo([]string{"--to", "123", "--file", "BAACAgIAAx0FakeFileID", "--duration", "soon"})
	if exitCode != 2 {
		t.Fatalf("cmdVideo exitCode = %d, want 2", exitCode)
	}
}

func TestCmdVideoRejectsInvalidSupportsStreaming(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVideo([]string{"--to", "123", "--file", "BAACAgIAAx0FakeFileID", "--supports-streaming=maybe"})
	if exitCode != 2 {
		t.Fatalf("cmdVideo exitCode = %d, want 2", exitCode)
	}
}
