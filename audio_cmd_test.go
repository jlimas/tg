package main

import "testing"

func TestCmdAudioRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdAudio([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdAudio exitCode = %d, want 2", exitCode)
	}
}

func TestCmdAudioRejectsInvalidDuration(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdAudio([]string{"--to", "123", "--file", "CQACAgIAAx0FakeFileID", "--duration", "soon"})
	if exitCode != 2 {
		t.Fatalf("cmdAudio exitCode = %d, want 2", exitCode)
	}
}
