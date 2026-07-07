package main

import "testing"

func TestCmdVoiceRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVoice([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdVoice exitCode = %d, want 2", exitCode)
	}
}

func TestCmdVoiceRejectsInvalidDuration(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdVoice([]string{"--to", "123", "--file", "AwACAgIAAx0FakeFileID", "--duration", "soon"})
	if exitCode != 2 {
		t.Fatalf("cmdVoice exitCode = %d, want 2", exitCode)
	}
}
