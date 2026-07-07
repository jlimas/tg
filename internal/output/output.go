// Package output renders tg's stdout in the compact, agent-friendly style
// described by the AXI conventions: key:value blocks, help hints, and
// structured errors, kept to a handful of fields.
package output

import (
	"fmt"
	"strings"
)

// KV is a single key/value line within a block.
type KV struct {
	Key   string
	Value string
}

// Block prints a named section followed by indented key: value lines, e.g.:
//
//	config:
//	  bot_token: (set)
//	  default_chat_id: 123456
func Block(name string, pairs []KV) {
	fmt.Println(name + ":")
	for _, p := range pairs {
		fmt.Printf("  %s: %s\n", p.Key, p.Value)
	}
}

// Help prints one or more suggested next-step commands:
//
//	help[2]:
//	  Run `tg text --to <chat_id> --message "..."` to send a message
//	  Run `tg config show` to see current configuration
func Help(lines ...string) {
	if len(lines) == 0 {
		return
	}
	if len(lines) == 1 {
		fmt.Printf("help[1]: %s\n", lines[0])
		return
	}
	fmt.Printf("help[%d]:\n", len(lines))
	for _, l := range lines {
		fmt.Printf("  %s\n", l)
	}
}

// Error prints a structured error to stdout, per AXI: errors are data, not
// noise, so they go on stdout in the same format as normal output.
func Error(msg string, help string) {
	fmt.Printf("error: %s\n", msg)
	if help != "" {
		fmt.Printf("help: %s\n", help)
	}
}

// Line prints a single top-level status line, e.g. "sent: message 42 to 123456".
func Line(format string, args ...any) {
	fmt.Println(strings.TrimSpace(fmt.Sprintf(format, args...)))
}
