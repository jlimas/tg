package app

import (
	"strings"
	"testing"
)

func TestCmdDiceRejectsInvalidEmoji(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := CmdDice([]string{"--to", "123", "--emoji", "🎃"})
		if exitCode != 2 {
			t.Fatalf("CmdDice exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, `invalid --emoji "🎃"`) {
		t.Fatalf("CmdDice output = %q, want invalid emoji message", output)
	}
	if !strings.Contains(output, "valid values: 🎲 🎯 🏀 ⚽ 🎳 🎰") {
		t.Fatalf("CmdDice output = %q, want valid emoji hint", output)
	}
}
