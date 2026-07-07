package cliflags

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseWithRepeatableCollectsOccurrences(t *testing.T) {
	values, multi, err := ParseWith([]string{
		"--tag", "one",
		"--tag=two",
		"--tag", "three",
	}, Spec{Repeatable: []string{"tag"}})
	if err != nil {
		t.Fatalf("ParseWith returned error: %v", err)
	}

	if len(values) != 0 {
		t.Fatalf("values = %#v, want empty", values)
	}
	want := []string{"one", "two", "three"}
	if !reflect.DeepEqual(multi["tag"], want) {
		t.Fatalf("multi[\"tag\"] = %#v, want %#v", multi["tag"], want)
	}
}

func TestParseWithBooleanBareDoesNotConsumeNextArg(t *testing.T) {
	values, multi, err := ParseWith([]string{"--silent", "hello"}, Spec{Boolean: []string{"silent"}})

	var unexpected *UnexpectedArgError
	if !errors.As(err, &unexpected) {
		t.Fatalf("err = %T %v, want *UnexpectedArgError", err, err)
	}
	if unexpected.Arg != "hello" {
		t.Fatalf("unexpected.Arg = %q, want %q", unexpected.Arg, "hello")
	}
	if values != nil {
		t.Fatalf("values = %#v, want nil on error", values)
	}
	if multi != nil {
		t.Fatalf("multi = %#v, want nil on error", multi)
	}
}

func TestParseWithBooleanSetsTrue(t *testing.T) {
	values, multi, err := ParseWith([]string{"--silent"}, Spec{Boolean: []string{"silent"}})
	if err != nil {
		t.Fatalf("ParseWith returned error: %v", err)
	}
	if values["silent"] != "true" {
		t.Fatalf("values[\"silent\"] = %q, want true", values["silent"])
	}
	if len(multi) != 0 {
		t.Fatalf("multi = %#v, want empty", multi)
	}
}

func TestParseWithBooleanExplicitValue(t *testing.T) {
	values, _, err := ParseWith([]string{"--silent=false"}, Spec{Boolean: []string{"silent"}})
	if err != nil {
		t.Fatalf("ParseWith returned error: %v", err)
	}
	if values["silent"] != "false" {
		t.Fatalf("values[\"silent\"] = %q, want false", values["silent"])
	}
}

func TestParseWithRepeatedPlainFlagErrors(t *testing.T) {
	_, _, err := ParseWith([]string{"--text", "hello", "--text=again"}, Spec{Allowed: []string{"text"}})

	var repeated *RepeatedFlagError
	if !errors.As(err, &repeated) {
		t.Fatalf("err = %T %v, want *RepeatedFlagError", err, err)
	}
	if repeated.Flag != "text" {
		t.Fatalf("repeated.Flag = %q, want %q", repeated.Flag, "text")
	}
}

func TestParseWithUnknownFlag(t *testing.T) {
	_, _, err := ParseWith([]string{"--bad", "value"}, Spec{Allowed: []string{"text"}})

	var unknown *UnknownFlagError
	if !errors.As(err, &unknown) {
		t.Fatalf("err = %T %v, want *UnknownFlagError", err, err)
	}
	if unknown.Flag != "bad" {
		t.Fatalf("unknown.Flag = %q, want %q", unknown.Flag, "bad")
	}
}

func TestParseWithMissingValue(t *testing.T) {
	_, _, err := ParseWith([]string{"--text", "--other"}, Spec{Allowed: []string{"text", "other"}})

	var missing *MissingValueError
	if !errors.As(err, &missing) {
		t.Fatalf("err = %T %v, want *MissingValueError", err, err)
	}
	if missing.Flag != "text" {
		t.Fatalf("missing.Flag = %q, want %q", missing.Flag, "text")
	}
}

func TestParseWithHelpAlwaysAccepted(t *testing.T) {
	values, multi, err := ParseWith([]string{"--help", "-h"}, Spec{})
	if err != nil {
		t.Fatalf("ParseWith returned error: %v", err)
	}
	if values["help"] != "true" {
		t.Fatalf("values[\"help\"] = %q, want true", values["help"])
	}
	if len(multi) != 0 {
		t.Fatalf("multi = %#v, want empty", multi)
	}
}

func TestParseExistingBehavior(t *testing.T) {
	t.Run("unknown flag", func(t *testing.T) {
		_, err := Parse([]string{"--bad", "value"}, []string{"text"})

		var unknown *UnknownFlagError
		if !errors.As(err, &unknown) {
			t.Fatalf("err = %T %v, want *UnknownFlagError", err, err)
		}
		if unknown.Flag != "bad" {
			t.Fatalf("unknown.Flag = %q, want %q", unknown.Flag, "bad")
		}
	})

	t.Run("missing value", func(t *testing.T) {
		_, err := Parse([]string{"--text", "--to"}, []string{"text", "to"})

		var missing *MissingValueError
		if !errors.As(err, &missing) {
			t.Fatalf("err = %T %v, want *MissingValueError", err, err)
		}
		if missing.Flag != "text" {
			t.Fatalf("missing.Flag = %q, want %q", missing.Flag, "text")
		}
	})

	t.Run("stray positional", func(t *testing.T) {
		_, err := Parse([]string{"hello"}, []string{"text"})

		var unexpected *UnexpectedArgError
		if !errors.As(err, &unexpected) {
			t.Fatalf("err = %T %v, want *UnexpectedArgError", err, err)
		}
		if unexpected.Arg != "hello" {
			t.Fatalf("unexpected.Arg = %q, want %q", unexpected.Arg, "hello")
		}
	})

	t.Run("help accepted", func(t *testing.T) {
		values, err := Parse([]string{"--help", "-h"}, nil)
		if err != nil {
			t.Fatalf("Parse returned error: %v", err)
		}
		if values["help"] != "true" {
			t.Fatalf("values[\"help\"] = %q, want true", values["help"])
		}
	})

	t.Run("repeat overwrites", func(t *testing.T) {
		values, err := Parse([]string{"--text", "first", "--text=second"}, []string{"text"})
		if err != nil {
			t.Fatalf("Parse returned error: %v", err)
		}
		if values["text"] != "second" {
			t.Fatalf("values[\"text\"] = %q, want second", values["text"])
		}
	})
}
