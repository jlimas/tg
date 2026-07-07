package main

import (
	"strings"
	"testing"
)

func TestCmdPollRequiresQuestion(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPoll([]string{"--to", "123", "--option", "Pizza", "--option", "Tacos"})
		if exitCode != 2 {
			t.Fatalf("cmdPoll exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--question is required") {
		t.Fatalf("cmdPoll output = %q, want --question required message", output)
	}
}

func TestCmdPollRejectsTooFewOptions(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPoll([]string{"--to", "123", "--question", "Lunch?", "--option", "Pizza"})
		if exitCode != 2 {
			t.Fatalf("cmdPoll exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--option must be given 2-10 times (got 1)") {
		t.Fatalf("cmdPoll output = %q, want too few options message", output)
	}
}

func TestCmdPollRejectsTooManyOptions(t *testing.T) {
	setTestConfigHome(t)

	args := []string{"--to", "123", "--question", "Pick one"}
	for _, option := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"} {
		args = append(args, "--option", option)
	}

	output := captureStdout(t, func() int {
		exitCode := cmdPoll(args)
		if exitCode != 2 {
			t.Fatalf("cmdPoll exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--option must be given 2-10 times (got 11)") {
		t.Fatalf("cmdPoll output = %q, want too many options message", output)
	}
}

func TestCmdPollQuizRequiresCorrectOption(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPoll([]string{"--to", "123", "--question", "2+2?", "--quiz", "--option", "3", "--option", "4"})
		if exitCode != 2 {
			t.Fatalf("cmdPoll exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--quiz requires --correct-option") {
		t.Fatalf("cmdPoll output = %q, want missing correct option message", output)
	}
}

func TestCmdPollRejectsCorrectOptionOutOfRange(t *testing.T) {
	setTestConfigHome(t)

	output := captureStdout(t, func() int {
		exitCode := cmdPoll([]string{"--to", "123", "--question", "2+2?", "--quiz", "--correct-option", "5", "--option", "3", "--option", "4"})
		if exitCode != 2 {
			t.Fatalf("cmdPoll exitCode = %d, want 2", exitCode)
		}
		return exitCode
	})

	if !strings.Contains(output, "--correct-option 5 is out of range (0-1 for 2 options)") {
		t.Fatalf("cmdPoll output = %q, want out of range message", output)
	}
}

func TestBuildPollOptionsJSONUsesInputPollOptionObjects(t *testing.T) {
	got, err := buildPollOptionsJSON([]string{"Pizza", "Tacos"})
	if err != nil {
		t.Fatalf("buildPollOptionsJSON returned error: %v", err)
	}

	want := `[{"text":"Pizza"},{"text":"Tacos"}]`
	if got != want {
		t.Fatalf("buildPollOptionsJSON = %q, want %q", got, want)
	}
}
