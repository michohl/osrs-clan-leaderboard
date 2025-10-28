package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// HiscoreCommandInfo is the information we'll use to
// "register" this command with Discord so it appears
// as an option to end users
var HiscoreCommandInfo = discordgo.ApplicationCommand{
	Name:        "hiscore",
	Description: "Get the hiscore(s) for a single user for specified activities",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:         "rsn",
			Description:  "The RSN of the user you want to get a hiscore of",
			Type:         discordgo.ApplicationCommandOptionString,
			Required:     true,
			Autocomplete: true,
		},
		{
			Name:        "activities",
			Description: "A comma separated list of activities to lookup",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

// HiscoreAutocompleteHandler is how we populate server tailored autocomplete options
func HiscoreAutocompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	choices := []*discordgo.ApplicationCommandOptionChoice{}

	allUsers, err := storage.FetchAllUsers(i.GuildID)
	if err != nil {
		log.Println(err)
		return
	}

	for _, u := range allUsers {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  u.OsrsUsername,
			Value: u.OsrsUsername,
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

	if err != nil {
		log.Println(err)
		return
	}
}

// HiscoreHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func HiscoreHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		hiscoreCommand(s, i)
	}
}

// Actually do the command the user is requesting
func hiscoreCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer our message so we have time to do processing
	// before discord times us out (we get 15 minutes now)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Generating hiscores message data...",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	data := i.ApplicationCommandData().Options
	osrsUsername := data[0].StringValue()
	activities := strings.Split(data[1].StringValue(), ",")

	discoveredErrors := ""

	osrsUser, err := storage.FetchUser(i.GuildID, hiscores.EncodeRSN(osrsUsername))
	if err != nil {
		log.Println(err)

		discoveredErrors = fmt.Sprintf(
			"%s\n* %s",
			discoveredErrors,
			fmt.Sprintf("Failed to find details of user '%s'. Make sure you have done /assign", osrsUsername),
		)
	}

	hiscoresEmbeds, err := hiscores.FormatEmbeds(activities, []model.Users{osrsUser})
	if err != nil {
		discoveredErrors = fmt.Sprintf(
			"%s\n* %s",
			discoveredErrors,
			err,
		)
	}

	if discoveredErrors != "" {

		toolsEmoji := types.ApplicationEmojis["tools"]
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: discoveredErrors,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: toolsEmoji.Name,
								ID:   toolsEmoji.ID,
							},
							Label: "List of All Skills and Activities",
							Style: discordgo.LinkButton,
							URL:   "https://runescape.wiki/w/Application_programming_interface#Old_School_Hiscores",
						},
					},
				},
			},
		})
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: hiscoresEmbeds,
	})
	if err != nil {
		log.Println(err)
		return
	}
}
