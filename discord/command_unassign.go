package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/storage"

	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
)

// UnassignCommandInfo builds an association between OSRS
// usernames and discord users
var UnassignCommandInfo = discordgo.ApplicationCommand{
	Name:        "unassign",
	Description: "Unassign an OSRS user to a discord member",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:         "osrs_user",
			Description:  "The user's OSRS username",
			Type:         discordgo.ApplicationCommandOptionString,
			Required:     true,
			Autocomplete: true,
		},
	},
}

// UnassignHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func UnassignHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		unassignCommand(s, i)
	}
}

// Actually do the command the user is requesting
func unassignCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	data := i.ApplicationCommandData().Options
	osrsUsername := data[0].StringValue()

	err := storage.RemoveUser(model.Users{
		OsrsUsernameKey: hiscores.EncodeRSN(osrsUsername),
		ServerID:        i.GuildID,
	})
	if err != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("OSRS User %s couldn't be unassigned...", osrsUsername),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("OSRS User %s unassigned", osrsUsername),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
}
