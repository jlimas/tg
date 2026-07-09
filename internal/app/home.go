package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
)

func CmdHome() int {
	fmt.Printf("bin: %s\n", binPath())
	fmt.Printf("description: %s\n", description)

	configured, err := config.Exists()
	if err != nil {
		output.Error(fmt.Sprintf("reading config: %v", err), "")
		return 1
	}

	if !configured {
		fmt.Println("config: not configured")
		output.Help(`tg config set --bot-token "<token>"`)
		return 0
	}

	fmt.Println("config: configured")
	output.Help(
		`tg text --to <chat_id> --message "..."`,
		`tg photo --to <chat_id> --file ./cat.jpg --caption "hi"`,
		`tg document --to <chat_id> --file ./report.pdf --caption "Q3"`,
		`tg video --to <chat_id> --file ./clip.mp4 --supports-streaming`,
		`tg media-group --to <chat_id> --file a.jpg --file b.jpg`,
		`tg paid-media --to <chat_id> --star-count 50 --file preview.jpg`,
		`tg video-note --to <chat_id> --file ./round.mp4`,
		`tg animation --to <chat_id> --file ./loop.gif --caption "launch"`,
		`tg sticker --to <chat_id> --file ./sticker.webp`,
		`tg audio --to <chat_id> --file ./song.mp3 --performer "X" --title "Y"`,
		`tg voice --to <chat_id> --file ./note.ogg`,
		`tg location --to <chat_id> --lat 40.7580 --long -73.9855`,
		`tg venue --to <chat_id> --lat 40.75 --long -73.98 --title "MSG" --address "4 Penn Plaza"`,
		`tg contact --to <chat_id> --phone "+15551234567" --first-name "Ada"`,
		`tg dice --to <chat_id> --emoji "🎯"`,
		`tg poll --to <chat_id> --question "Lunch?" --option Pizza --option Tacos`,
		"tg listen",
		"tg config show",
	)
	return 0
}

// binPath returns the path to the running executable, with the user's home
// directory collapsed to ~, per AXI's home-view convention.
func binPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "tg"
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(exe, home) {
		return "~" + strings.TrimPrefix(exe, home)
	}
	return exe
}
