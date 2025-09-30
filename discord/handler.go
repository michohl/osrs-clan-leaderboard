package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	&HelpCommandInfo,
	&PingCommandInfo,
	&ConfigureCommandInfo,
}

// CommandHandler is the contract any function we want to use as a handler must satisfy
type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

var commandHandlers = map[string]CommandHandler{
	"help":      HelpHandler,
	"ping":      PingHandler,
	"configure": ConfigureHandler,
}

// GetCommandHandler takes the user specified command and returns
// the relevant function responsible for responding to the user
func GetCommandHandler(command string) CommandHandler {
	if h, ok := commandHandlers[command]; ok {
		return h
	}

	log.Fatalf("NO Command Handler defined for %s", command)
	return nil
}

// GetModalSubmitHandler takes the custom ID of a modal survey
// and tries to find which function should handle the submit action
// for it. Each modal survey is prefixed with a hardcoded string
// we will use as an identifier. We can't use a literal map because
// each modal survey is suffixed with the user's ID to make it unique.
func GetModalSubmitHandler(customID string) CommandHandler {
	switch {
	case strings.Contains(customID, "modals_survey_configure_"):
		return ConfigureModalSubmit
	default:
		log.Fatalf("No Modal Submit Handler that matches %s", customID)
	}

	return nil
}
