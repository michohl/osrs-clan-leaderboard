package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/types"
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

	documentationEmoji := types.ApplicationEmojis["documents"]
	githubEmoji := types.ApplicationEmojis["github"]

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "TBD",
			Flags:   discordgo.MessageFlagsEphemeral,
			// Embeds:  []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: documentationEmoji.Name,
								ID:   documentationEmoji.ID,
							},
							Label: "Documentation",
							Style: discordgo.LinkButton,
							URL:   "https://github.com/michohl/osrs-clan-leaderboard",
						},
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: githubEmoji.Name,
								ID:   githubEmoji.ID,
							},
							Label: "Report an Issue",
							Style: discordgo.LinkButton,
							URL:   "https://github.com/michohl/osrs-clan-leaderboard/issues",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
}
