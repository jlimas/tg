package main

import (
	"fmt"

	"github.com/jlimas/tg/internal/output"
)

const description = "tg — send Telegram messages from the command line"

func cmdHelp() int {
	fmt.Println(description)
	fmt.Println()
	fmt.Println("usage: tg <command> [flags]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  config set   save the bot token and optional default chat id")
	fmt.Println("  config show  show the current configuration")
	fmt.Println("  text         send a text message")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println(`  tg config set --bot-token "123:ABC..."`)
	fmt.Println(`  tg text --to 123456789 --message "hello"`)
	fmt.Println()
	fmt.Println("run `tg <command> --help` for details on a specific command")
	return 0
}

func unknownCommand(name string) int {
	output.Error(fmt.Sprintf("unknown command %q", name), "tg --help")
	return 2
}
