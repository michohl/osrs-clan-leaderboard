package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// PingCommandInfo is the information we'll use to
// "register" this command with Discord so it appears
// as an option to end users
var PingCommandInfo = discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Test if the bot is up and running",
	Type:        discordgo.ChatApplicationCommand,
	Options:     []*discordgo.ApplicationCommandOption{},
}

// PingHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func PingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		pingCommand(s, i)
	}
}

// Actually do the command the user is requesting
func pingCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "pong",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
}
