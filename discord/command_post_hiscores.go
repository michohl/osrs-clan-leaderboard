package discord

import (
	"github.com/bwmarrin/discordgo"
)

// PostHiscoresCommandInfo is the information we'll use to
// "register" this command with Discord so it appears
// as an option to end users
var PostHiscoresCommandInfo = discordgo.ApplicationCommand{
	Name:        "post",
	Description: "Manually invoke posting a hiscores message",
	Type:        discordgo.ChatApplicationCommand,
	Options:     []*discordgo.ApplicationCommandOption{},
}

// PostHiscoresHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func PostHiscoresHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		postHiscoresCommand(s, i)
	}
}

// Actually do the command the user is requesting
func postHiscoresCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Posting the message takes longer than our timeout allows so we'll just
	// tell the user we're taking action then continue processing in the background
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Messaged being generated now...",
		},
	})
	if err != nil {
		panic(err)
	}

	err = PostHiscoresMessage(i.GuildID, s)
	if err != nil {
		panic(err)
	}

}
