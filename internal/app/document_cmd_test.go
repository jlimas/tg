package app

import "testing"

func TestCmdDocumentRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := CmdDocument([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("CmdDocument exitCode = %d, want 2", exitCode)
	}
}
