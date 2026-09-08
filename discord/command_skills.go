package discord

import (
	"bytes"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/skills"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// SkillsCommandInfo is the information we'll use to
// "register" this command with Discord so it appears
// as an option to end users
var SkillsCommandInfo = discordgo.ApplicationCommand{
	Name:        "skills",
	Description: "Get the skills page for a single user as it appears in game",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:         "rsn",
			Description:  "The RSN of the user you want to get the skills page of",
			Type:         discordgo.ApplicationCommandOptionString,
			Required:     true,
			Autocomplete: true,
		},
		{
			Name:        "leaderboard",
			Description: "Which leaderboard you want to check hiscores on",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Use Default for Acccount Type",
					Value: "",
				},
				{
					Name:  "Main",
					Value: "main",
				},
				{
					Name:  "Ironman",
					Value: "ironman",
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
					Name:  "Seasonal (Leagues/DMM)",
					Value: "seasonal",
				},
			},
		},
	},
}

// SkillsAutocompleteHandler is how we populate server tailored autocomplete options
func SkillsAutocompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

// SkillsHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func SkillsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		skillsCommand(s, i)
	}
}

// Actually do the command the user is requesting
func skillsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer our message so we have time to do processing
	// before discord times us out (we get 15 minutes now)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Generating skills page data...",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	data := i.ApplicationCommandData().Options
	osrsUsername := data[0].StringValue()

	accountType := ""
	overrideLeaderboard := false
	if len(data) > 1 {
		overrideLeaderboard = true
		accountType = data[1].StringValue()
	}

	osrsUser, err := storage.FetchUser(i.GuildID, hiscores.EncodeRSN(osrsUsername))
	if err != nil {
		log.Println("Failed to find user details. Attempting to make ad-hoc details")
		osrsUser = model.Users{
			OsrsUsernameKey: hiscores.EncodeRSN(osrsUsername),
			OsrsUsername:    osrsUsername,
			ServerID:        i.GuildID,
			OsrsAccountType: accountType,
			DiscordUsername: "",
			DiscordUserID:   "",
		}
	}

	if overrideLeaderboard {
		osrsUser.OsrsAccountType = accountType
	} else {
		log.Println("No account type provided. Discovering account type")
		osrsUser.OsrsAccountType = hiscores.DiscoverRSAccountType(osrsUsername)
	}

	discoveredErrors := ""

	userHiscores, err := hiscores.GetUserHiscores([]model.Users{osrsUser}, "")
	if len(userHiscores) == 0 {
		err := fmt.Errorf("Unable to find hiscores for user '%s' on '%s' leaderboard", osrsUsername, osrsUser.OsrsAccountType)
		discoveredErrors = fmt.Sprintf(
			"%s\n* %s",
			discoveredErrors,
			err,
		)
	}

	// Render users skills as they appear in game
	var imgBytes []byte
	for _, hiscores := range userHiscores {
		imgBytes, err = skills.RenderSkillsPage(osrsUser, &hiscores)
		if err != nil || len(imgBytes) == 0 {
			log.Println(err)
			discoveredErrors = fmt.Sprintf(
				"%s\n* Failed to render skills page: %s",
				discoveredErrors,
				err,
			)
		}
	}

	// Get the emoji for the account type
	accountTypeEmoji, ok := types.ApplicationEmojis[osrsUser.OsrsAccountType]
	accountEmoji := ""
	if ok {
		accountEmoji = fmt.Sprintf(" <:%s>", accountTypeEmoji.APIName())
	}

	// If this is a user that is assigned in the server then we should tag the associated discord user
	// If this was an ad-hoc command for an unassigned user then just omit the discord user tag
	discordUser := ""
	if osrsUser.DiscordUserID != "" {
		discordUser = fmt.Sprintf(" (<@%s>)", osrsUser.DiscordUserID)
	}

	// Build the WebhookParams payload with the file attachment
	params := &discordgo.WebhookParams{
		Files: []*discordgo.File{
			{
				Name:   "skills.png",
				Reader: bytes.NewReader(imgBytes),
			},
		},
		Content: fmt.Sprintf("Here is the skills page for%s %s%s:", accountEmoji, osrsUser.OsrsUsername, discordUser),
	}

	if discoveredErrors != "" {
		githubEmoji := types.ApplicationEmojis["github"]
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: discoveredErrors,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
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
		})
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	// Send the followup message with the rendered image
	_, err = s.FollowupMessageCreate(i.Interaction, true, params)
	if err != nil {
		log.Println(err)
		return
	}
}
