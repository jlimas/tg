package main

import (
	"testing"

	"github.com/jlimas/tg/internal/cliflags"
)

func TestCommonBooleanFlagsFromParseWith(t *testing.T) {
	spec := cliflags.Spec{
		Allowed: append([]string{}, commonAllowedFlagNames...),
		Boolean: append([]string{}, commonBooleanFlagNames...),
	}

	values, _, err := cliflags.ParseWith([]string{"--silent"}, spec)
	if err != nil {
		t.Fatalf("ParseWith bare --silent returned error: %v", err)
	}
	params, exitCode := commonParamsFrom(values, "123", textUsage)
	if exitCode != 0 {
		t.Fatalf("commonParamsFrom bare --silent exitCode = %d, want 0", exitCode)
	}
	if !params.DisableNotification {
		t.Fatal("DisableNotification = false, want true")
	}

	values, _, err = cliflags.ParseWith([]string{"--silent=false"}, spec)
	if err != nil {
		t.Fatalf("ParseWith --silent=false returned error: %v", err)
	}
	params, exitCode = commonParamsFrom(values, "123", textUsage)
	if exitCode != 0 {
		t.Fatalf("commonParamsFrom --silent=false exitCode = %d, want 0", exitCode)
	}
	if params.DisableNotification {
		t.Fatal("DisableNotification = true, want false")
	}

	values, _, err = cliflags.ParseWith([]string{"--silent=maybe"}, spec)
	if err != nil {
		t.Fatalf("ParseWith --silent=maybe returned error: %v", err)
	}
	_, exitCode = commonParamsFrom(values, "123", textUsage)
	if exitCode != 2 {
		t.Fatalf("commonParamsFrom --silent=maybe exitCode = %d, want 2", exitCode)
	}
}
