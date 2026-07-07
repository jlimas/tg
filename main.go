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
	case "text":
		return cmdText(args[1:])
	case "photo":
		return cmdPhoto(args[1:])
	case "document":
		return cmdDocument(args[1:])
	case "video":
		return cmdVideo(args[1:])
	case "media-group":
		return cmdMediaGroup(args[1:])
	case "paid-media":
		return cmdPaidMedia(args[1:])
	case "video-note":
		return cmdVideoNote(args[1:])
	case "animation":
		return cmdAnimation(args[1:])
	case "sticker":
		return cmdSticker(args[1:])
	case "audio":
		return cmdAudio(args[1:])
	case "voice":
		return cmdVoice(args[1:])
	case "location":
		return cmdLocation(args[1:])
	case "venue":
		return cmdVenue(args[1:])
	case "contact":
		return cmdContact(args[1:])
	case "dice":
		return cmdDice(args[1:])
	case "poll":
		return cmdPoll(args[1:])
	default:
		return unknownCommand(args[0])
	}
}
