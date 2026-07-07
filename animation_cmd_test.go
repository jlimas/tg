package main

import "testing"

func TestCmdAnimationRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdAnimation([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdAnimation exitCode = %d, want 2", exitCode)
	}
}

func TestCmdAnimationRejectsInvalidDuration(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdAnimation([]string{"--to", "123", "--file", "CgACAgIAAx0FakeFileID", "--duration", "soon"})
	if exitCode != 2 {
		t.Fatalf("cmdAnimation exitCode = %d, want 2", exitCode)
	}
}

func TestCmdAnimationRejectsInvalidSpoiler(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdAnimation([]string{"--to", "123", "--file", "CgACAgIAAx0FakeFileID", "--spoiler=maybe"})
	if exitCode != 2 {
		t.Fatalf("cmdAnimation exitCode = %d, want 2", exitCode)
	}
}
