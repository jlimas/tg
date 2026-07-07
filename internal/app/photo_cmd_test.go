package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmdPhotoRequiresFile(t *testing.T) {
	setTestConfigHome(t)

	exitCode := CmdPhoto([]string{"--to", "123"})
	if exitCode != 2 {
		t.Fatalf("CmdPhoto exitCode = %d, want 2", exitCode)
	}
}

func TestCmdPhotoRejectsInvalidSpoiler(t *testing.T) {
	setTestConfigHome(t)

	exitCode := CmdPhoto([]string{"--to", "123", "--file", "AgACAgIAAx0FakeFileID", "--spoiler=maybe"})
	if exitCode != 2 {
		t.Fatalf("CmdPhoto exitCode = %d, want 2", exitCode)
	}
}

func setTestConfigHome(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	configDir := filepath.Join(home, ".config", "tg")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("make config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	config := "bot_token = \"123:ABC\"\ndefault_chat_id = \"123\"\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}
