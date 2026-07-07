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
