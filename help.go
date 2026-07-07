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
	fmt.Println("  media-group  send a media album")
	fmt.Println("  paid-media   send paid media")
	fmt.Println("  video-note   send a video note")
	fmt.Println("  animation    send an animation")
	fmt.Println("  sticker      send a sticker")
	fmt.Println("  audio        send an audio file")
	fmt.Println("  voice        send a voice message")
	fmt.Println("  location     send a location")
	fmt.Println("  venue        send a venue")
	fmt.Println("  contact      send a contact")
	fmt.Println("  dice         send a dice animation")
	fmt.Println("  poll         send a poll")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println(`  tg config set --bot-token "123:ABC..."`)
	fmt.Println(`  tg text --to 123456789 --message "hello"`)
	fmt.Println(`  tg photo --to 123456789 --file ./cat.jpg --caption "hi"`)
	fmt.Println(`  tg document --to 123456789 --file ./report.pdf --caption "Q3"`)
	fmt.Println(`  tg video --to 123456789 --file ./clip.mp4 --supports-streaming`)
	fmt.Println(`  tg media-group --to 123456789 --file a.jpg --file b.jpg`)
	fmt.Println(`  tg paid-media --to 123456789 --star-count 50 --file preview.jpg`)
	fmt.Println(`  tg video-note --to 123456789 --file ./round.mp4`)
	fmt.Println(`  tg animation --to 123456789 --file ./loop.gif --caption "launch"`)
	fmt.Println(`  tg sticker --to 123456789 --file ./sticker.webp`)
	fmt.Println(`  tg audio --to 123456789 --file ./song.mp3 --performer "X" --title "Y"`)
	fmt.Println(`  tg voice --to 123456789 --file ./note.ogg`)
	fmt.Println(`  tg location --to 123456789 --lat 40.7580 --long -73.9855`)
	fmt.Println(`  tg venue --to 123456789 --lat 40.75 --long -73.98 --title "MSG" --address "4 Penn Plaza"`)
	fmt.Println(`  tg contact --to 123456789 --phone "+15551234567" --first-name "Ada"`)
	fmt.Println(`  tg dice --to 123456789 --emoji "🎯"`)
	fmt.Println(`  tg poll --to 123456789 --question "Lunch?" --option Pizza --option Tacos`)
	fmt.Println()
	fmt.Println("run `tg <command> --help` for details on a specific command")
	return 0
}

func unknownCommand(name string) int {
	output.Error(fmt.Sprintf("unknown command %q", name), "tg --help")
	return 2
}
