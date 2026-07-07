// Command tg is a small AXI-style CLI for sending Telegram messages
// through a bot, configured via ~/.config/tg/config.toml.
package main

import (
	"os"

	"github.com/jlimas/tg/internal/app"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		return app.CmdHome()
	}

	switch args[0] {
	case "--help", "-h":
		return app.CmdHelp()
	case "config":
		return app.DispatchConfig(args[1:])
	case "text":
		return app.CmdText(args[1:])
	case "photo":
		return app.CmdPhoto(args[1:])
	case "document":
		return app.CmdDocument(args[1:])
	case "video":
		return app.CmdVideo(args[1:])
	case "media-group":
		return app.CmdMediaGroup(args[1:])
	case "paid-media":
		return app.CmdPaidMedia(args[1:])
	case "video-note":
		return app.CmdVideoNote(args[1:])
	case "animation":
		return app.CmdAnimation(args[1:])
	case "sticker":
		return app.CmdSticker(args[1:])
	case "audio":
		return app.CmdAudio(args[1:])
	case "voice":
		return app.CmdVoice(args[1:])
	case "location":
		return app.CmdLocation(args[1:])
	case "venue":
		return app.CmdVenue(args[1:])
	case "contact":
		return app.CmdContact(args[1:])
	case "dice":
		return app.CmdDice(args[1:])
	case "poll":
		return app.CmdPoll(args[1:])
	default:
		return app.UnknownCommand(args[0])
	}
}
