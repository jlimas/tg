package app

import (
	"fmt"

	"github.com/jlimas/tg/internal/cliflags"
	"github.com/jlimas/tg/internal/config"
	"github.com/jlimas/tg/internal/output"
)

func DispatchConfig(args []string) int {
	if len(args) == 0 {
		output.Error("missing config subcommand", "tg config --help")
		return 2
	}

	switch args[0] {
	case "--help", "-h":
		return configHelp()
	case "set":
		return cmdConfigSet(args[1:])
	case "show":
		return cmdConfigShow(args[1:])
	default:
		output.Error(fmt.Sprintf("unknown config subcommand %q", args[0]), "valid subcommands: set, show")
		return 2
	}
}

func configHelp() int {
	fmt.Println("tg config — manage ~/.config/tg/config.toml")
	fmt.Println()
	fmt.Println("subcommands:")
	fmt.Println("  set --bot-token <token> [--default-chat-id <id>]")
	fmt.Println("  show")
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println(`  tg config set --bot-token "123456:AAExample-Token"`)
	fmt.Println(`  tg config set --default-chat-id 987654321`)
	fmt.Println("  tg config show")
	return 0
}

func cmdConfigSet(args []string) int {
	values, err := cliflags.Parse(args, []string{"bot-token", "default-chat-id"})
	if err != nil {
		return flagError(err, "tg config set --bot-token <token> [--default-chat-id <id>]")
	}
	if values["help"] == "true" {
		fmt.Println("tg config set — save the bot token and/or default chat id")
		fmt.Println()
		fmt.Println("flags:")
		fmt.Println("  --bot-token <token>        bot token from @BotFather")
		fmt.Println("  --default-chat-id <id>     chat id used by `tg text` when --to is omitted")
		fmt.Println()
		fmt.Println("example:")
		fmt.Println(`  tg config set --bot-token "123456:AAExample-Token" --default-chat-id 987654321`)
		return 0
	}

	botToken, hasBotToken := values["bot-token"]
	defaultChatID, hasDefaultChatID := values["default-chat-id"]
	if !hasBotToken && !hasDefaultChatID {
		output.Error("at least one of --bot-token or --default-chat-id is required", "tg config set --bot-token <token> [--default-chat-id <id>]")
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		output.Error(fmt.Sprintf("reading existing config: %v", err), "")
		return 1
	}
	if hasBotToken {
		cfg.BotToken = botToken
	}
	if hasDefaultChatID {
		cfg.DefaultChatID = defaultChatID
	}
	if cfg.BotToken == "" {
		output.Error("bot token is required on first setup", `tg config set --bot-token "<token>"`)
		return 2
	}

	if err := config.Save(cfg); err != nil {
		output.Error(fmt.Sprintf("saving config: %v", err), "")
		return 1
	}

	path, _ := config.Path()
	output.Line("config: saved to %s", path)
	output.Help(
		"tg config show",
		`tg text --to <chat_id> --message "..."`,
	)
	return 0
}

func cmdConfigShow(args []string) int {
	values, err := cliflags.Parse(args, nil)
	if err != nil {
		return flagError(err, "tg config show")
	}
	if values["help"] == "true" {
		fmt.Println("tg config show — print the current configuration (bot token masked)")
		return 0
	}

	exists, err := config.Exists()
	if err != nil {
		output.Error(fmt.Sprintf("reading config: %v", err), "")
		return 1
	}
	if !exists {
		fmt.Println("config: not configured")
		output.Help(`tg config set --bot-token "<token>"`)
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		output.Error(fmt.Sprintf("reading config: %v", err), "")
		return 1
	}
	path, _ := config.Path()

	tokenDisplay := "(not set)"
	if cfg.BotToken != "" {
		tokenDisplay = maskToken(cfg.BotToken)
	}
	chatDisplay := "(not set)"
	if cfg.DefaultChatID != "" {
		chatDisplay = cfg.DefaultChatID
	}

	output.Block("config", []output.KV{
		{Key: "path", Value: path},
		{Key: "bot_token", Value: tokenDisplay},
		{Key: "default_chat_id", Value: chatDisplay},
	})
	return 0
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "(set)"
	}
	return token[:4] + "…" + token[len(token)-4:]
}
