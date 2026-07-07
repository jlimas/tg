// Package cliflags is a tiny --flag/--flag=value parser shared by tg's
// subcommands. It exists (instead of the stdlib flag package) so unknown
// flags and stray positional arguments are rejected explicitly, per AXI's
// "fail loud on unrecognized input" rule.
package cliflags

import "strings"

// UnknownFlagError is returned when an argument starting with "--" is not
// in the caller's allowed set.
type UnknownFlagError struct {
	Flag string
}

func (e *UnknownFlagError) Error() string {
	return "unknown flag --" + e.Flag
}

// UnexpectedArgError is returned for a bare positional argument, which none
// of tg's commands currently accept.
type UnexpectedArgError struct {
	Arg string
}

func (e *UnexpectedArgError) Error() string {
	return "unexpected argument " + e.Arg
}

// MissingValueError is returned when a flag is the last argument, or is
// immediately followed by another flag, leaving it without a value.
type MissingValueError struct {
	Flag string
}

func (e *MissingValueError) Error() string {
	return "flag --" + e.Flag + " requires a value"
}

// RepeatedFlagError is returned when a flag appears more than once without
// being declared repeatable by the caller.
type RepeatedFlagError struct {
	Flag string
}

func (e *RepeatedFlagError) Error() string {
	return "flag --" + e.Flag + " was already set"
}

// Spec declares the flags accepted by ParseWith.
type Spec struct {
	Allowed    []string
	Repeatable []string
	Boolean    []string
}

type flagKind int

const (
	flagPlain flagKind = iota + 1
	flagRepeatable
	flagBoolean
)

// Parse reads --name value and --name=value pairs from args. "--help" and
// "-h" are always accepted and reported back as values["help"] = "true",
// regardless of allowed. Any other "--name" not present in allowed produces
// an *UnknownFlagError. A bare positional argument produces an
// *UnexpectedArgError.
func Parse(args []string, allowed []string) (map[string]string, error) {
	allowedSet := make(map[string]bool, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = true
	}

	values := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			values["help"] = "true"
			continue
		}

		if !strings.HasPrefix(arg, "--") {
			return nil, &UnexpectedArgError{Arg: arg}
		}

		name, val, hasEq := strings.Cut(arg[2:], "=")
		if !allowedSet[name] {
			return nil, &UnknownFlagError{Flag: name}
		}
		if !hasEq {
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				return nil, &MissingValueError{Flag: name}
			}
			val = args[i+1]
			i++
		}
		values[name] = val
	}
	return values, nil
}

// ParseWith reads args using spec. Plain flags behave like Parse: they require
// a value and may only appear once. Repeatable flags require values and collect
// every occurrence in order in the returned multi map. Boolean flags set
// values[name] to "true" when present without consuming the next argument, and
// also accept --name=true or --name=false.
func ParseWith(args []string, spec Spec) (values map[string]string, multi map[string][]string, err error) {
	flags := make(map[string]flagKind, len(spec.Allowed)+len(spec.Repeatable)+len(spec.Boolean))
	for _, a := range spec.Allowed {
		flags[a] = flagPlain
	}
	for _, a := range spec.Repeatable {
		flags[a] = flagRepeatable
	}
	for _, a := range spec.Boolean {
		flags[a] = flagBoolean
	}

	values = make(map[string]string)
	multi = make(map[string][]string)
	seen := make(map[string]bool)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			values["help"] = "true"
			continue
		}

		if !strings.HasPrefix(arg, "--") {
			return nil, nil, &UnexpectedArgError{Arg: arg}
		}

		name, val, hasEq := strings.Cut(arg[2:], "=")
		kind, ok := flags[name]
		if !ok {
			return nil, nil, &UnknownFlagError{Flag: name}
		}

		if kind != flagRepeatable && seen[name] {
			return nil, nil, &RepeatedFlagError{Flag: name}
		}
		seen[name] = true

		if kind == flagBoolean {
			if !hasEq {
				val = "true"
			}
			values[name] = val
			continue
		}

		if !hasEq {
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				return nil, nil, &MissingValueError{Flag: name}
			}
			val = args[i+1]
			i++
		}

		if kind == flagRepeatable {
			multi[name] = append(multi[name], val)
			continue
		}
		values[name] = val
	}
	return values, multi, nil
}
