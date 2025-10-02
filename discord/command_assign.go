package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// AssignCommandInfo builds an association between OSRS
// usernames and discord users
var AssignCommandInfo = discordgo.ApplicationCommand{
	Name:        "assign",
	Description: "Assign an OSRS user to a discord member",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "discord_user",
			Description: "The user's Discord handle",
			Type:        discordgo.ApplicationCommandOptionUser,
			Required:    true,
		},
		{
			Name:        "osrs_user",
			Description: "The user's OSRS username",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "osrs_account_type",
			Description: "The 'kind' of Account (all kinds will use the global leaderboard)",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Main",
					Value: "main",
				},
				{
					Name:  "Ironman",
					Value: "ironman",
				},
				{
					Name:  "Unranked Group Ironman",
					Value: "unranked_group_ironman",
				},
				{
					Name:  "Group Ironman",
					Value: "group_ironman",
				},
				{
					Name:  "Ultimate Ironman",
					Value: "ultimate_ironman",
				},
				{
					Name:  "Hardcore Ironman",
					Value: "hardcore_ironman",
				},
				{
					Name:  "Hardcore Group Ironman",
					Value: "hardcore_group_ironman",
				},
			},
		},
	},
}

// AssignHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func AssignHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		assignCommand(s, i)
	}
}

// Actually do the command the user is requesting
func assignCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	data := i.ApplicationCommandData().Options
	discordUser := data[0].UserValue(s)
	osrsUsername := data[1].StringValue()
	osrsAccountType := data[2].StringValue()

	_, err := hiscores.GetPlayerHiscores(types.OSRSUser{Username: osrsUsername, AccountType: osrsAccountType})
	if err != nil {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("OSRS User %s couldn't be found...", osrsUsername),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	err = storage.EnrollUser(
		i.GuildID,
		discordUser,
		types.OSRSUser{
			Username:    osrsUsername,
			AccountType: osrsAccountType,
		},
	)
	if err != nil {
		log.Println(err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("OSRS User %s assigned to <@%s>", osrsUsername, discordUser.ID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
}
