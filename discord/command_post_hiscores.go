package discord

import (
	"log"

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
	// Defer our message so we have time to do processing
	// before discord times us out (we get 15 minutes now)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Generating hiscores message data...",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	err = PostHiscoresMessage(i.GuildID, s)
	if err != nil {
		log.Println(err)

		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Failed to post hiscores message...",
		})
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Flags:   discordgo.MessageFlagsEphemeral,
		Content: "Hiscores message posted!",
	})
	if err != nil {
		log.Println(err)
		return
	}
}
