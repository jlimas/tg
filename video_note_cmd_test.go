package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestCmdVideoNoteRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVideoNote([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdVideoNote exitCode = %d, want 2", exitCode)
	}
}

func TestCmdVideoNoteRejectsRemoteURLFile(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdVideoNote([]string{"--to", "123", "--file", "https://example.com/round.mp4"})
		if exitCode != 2 {
			t.Fatalf("cmdVideoNote exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "video notes don't support URLs") {
		t.Fatalf("cmdVideoNote output = %q, want URL rejection message", output)
	}
	if !strings.Contains(output, "download the file locally or use a previously uploaded file_id") {
		t.Fatalf("cmdVideoNote output = %q, want local download/file_id hint", output)
	}
}

func TestCmdVideoNoteRejectsInvalidDuration(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVideoNote([]string{"--to", "123", "--file", "DQACAgIAAx0FakeFileID", "--duration", "soon"})
	if exitCode != 2 {
		t.Fatalf("cmdVideoNote exitCode = %d, want 2", exitCode)
	}
}

func TestCmdVideoNoteRejectsInvalidLength(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVideoNote([]string{"--to", "123", "--file", "DQACAgIAAx0FakeFileID", "--length", "wide"})
	if exitCode != 2 {
		t.Fatalf("cmdVideoNote exitCode = %d, want 2", exitCode)
	}
}

func captureStdout(t *testing.T, fn func() int) string {
	t.Helper()

	oldStdout := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = write
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	fn()

	if err := write.Close(); err != nil {
		t.Fatalf("close stdout pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	data, err := io.ReadAll(read)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := read.Close(); err != nil {
		t.Fatalf("close stdout pipe reader: %v", err)
	}
	return string(data)
}
