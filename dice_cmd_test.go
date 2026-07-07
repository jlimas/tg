package main

import (
	"strings"
	"testing"
)

func TestCmdDiceRejectsInvalidEmoji(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdDice([]string{"--to", "123", "--emoji", "🎃"})
		if exitCode != 2 {
			t.Fatalf("cmdDice exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, `invalid --emoji "🎃"`) {
		t.Fatalf("cmdDice output = %q, want invalid emoji message", output)
	}
	if !strings.Contains(output, "valid values: 🎲 🎯 🏀 ⚽ 🎳 🎰") {
		t.Fatalf("cmdDice output = %q, want valid emoji hint", output)
	}
}
