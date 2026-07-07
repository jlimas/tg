package main

import "testing"

func TestCmdStickerRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdSticker([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("cmdSticker exitCode = %d, want 2", exitCode)
	}
}

func TestCmdStickerRejectsCaption(t *testing.T) {
	setTestConfigHome(t)

	exitCode := cmdSticker([]string{"--to", "123", "--file", "CAACAgIAAx0FakeFileID", "--caption", "hi"})
	if exitCode != 2 {
		t.Fatalf("cmdSticker exitCode = %d, want 2", exitCode)
	}
}
