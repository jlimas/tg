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
	fmt.Println("  photo        send a photo")
	fmt.Println("  document     send a document")
	fmt.Println("  video        send a video")
	fmt.Println("  audio        send an audio file")
	fmt.Println("  voice        send a voice message")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println(`  tg config set --bot-token "123:ABC..."`)
	fmt.Println(`  tg text --to 123456789 --message "hello"`)
	fmt.Println(`  tg photo --to 123456789 --file ./cat.jpg --caption "hi"`)
	fmt.Println(`  tg document --to 123456789 --file ./report.pdf --caption "Q3"`)
	fmt.Println(`  tg video --to 123456789 --file ./clip.mp4 --supports-streaming`)
	fmt.Println(`  tg audio --to 123456789 --file ./song.mp3 --performer "X" --title "Y"`)
	fmt.Println(`  tg voice --to 123456789 --file ./note.ogg`)
	fmt.Println()
	fmt.Println("run `tg <command> --help` for details on a specific command")
	return 0
}

func unknownCommand(name string) int {
	output.Error(fmt.Sprintf("unknown command %q", name), "tg --help")
	return 2
}
