package discord

import (
	"github.com/bwmarrin/discordgo"
)

// HelpCommandInfo is the information we'll use to
// "register" this command with Discord so it appears
// as an option to end users
var HelpCommandInfo = discordgo.ApplicationCommand{
	Name:        "help",
	Description: "Give instructions on using this bot",
	Type:        discordgo.ChatApplicationCommand,
	Options:     []*discordgo.ApplicationCommandOption{},
}

// HelpHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func HelpHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		helpCommand(s, i)
	}
}

// Actually do the command the user is requesting
func helpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "TBD",
			/*
				// TODO: Figure out how to make a pretty help message
				Components: []discordgo.MessageComponent{
					discordgo.Container{
						Spoiler: false,
						Components: []discordgo.MessageComponent{
							discordgo.TextDisplay{
								Content: "# This is a test\nDetails here",
							},
							discordgo.TextDisplay{
								Content: "# This is a second test\nOther details here",
							},
						},
					},
				},
			*/
		},
	})
	if err != nil {
		panic(err)
	}
}
