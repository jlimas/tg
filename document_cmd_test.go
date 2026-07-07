package main

import "testing"

func TestCmdDocumentRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdDocument([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdDocument exitCode = %d, want 2", exitCode)
	}
}
