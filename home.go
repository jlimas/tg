package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
)

func cmdHome() int {
	fmt.Printf("bin: %s\n", binPath())
	fmt.Printf("description: %s\n", description)

	configured, err := config.Exists()
	if err != nil {
		output.Error(fmt.Sprintf("reading config: %v", err), "")
		return 1
	}

	if !configured {
		fmt.Println("config: not configured")
		output.Help(`tg config set --bot-token "<token>"`)
		return 0
	}

	fmt.Println("config: configured")
	output.Help(
		`tg send --to <chat_id> --text "..."`,
		"tg config show",
	)
	return 0
}

// binPath returns the path to the running executable, with the user's home
// directory collapsed to ~, per AXI's home-view convention.
func binPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "tg"
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(exe, home) {
		return "~" + strings.TrimPrefix(exe, home)
	}
	return exe
}
