// Command tg is a small AXI-style CLI for sending Telegram messages
// through a bot, configured via ~/.config/tg/config.toml.
package main

import "os"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		return cmdHome()
	}

	switch args[0] {
	case "--help", "-h":
		return cmdHelp()
	case "config":
		return dispatchConfig(args[1:])
	case "send":
		return cmdSend(args[1:])
	default:
		return unknownCommand(args[0])
	}
}
