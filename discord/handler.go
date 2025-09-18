package discord

import (
	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	&HelpCommandInfo,
	&PingCommandInfo,
}

// CommandHandler is the contract any function we want to use as a handler must satisfy
type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

var commandHandlers = map[string]CommandHandler{
	"help": HelpHandler,
	"ping": PingHandler,
}
