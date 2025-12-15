package discord

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/schedule"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
	"github.com/michohl/osrs-clan-leaderboard/utils"
)

// ConfigureCommandInfo is the information we'll use to
// "register" this command with Discord so it appears
// as an option to end users
var ConfigureCommandInfo = discordgo.ApplicationCommand{
	Name:        "configure",
	Description: "Allow the user to configure the server",
	//Type:        discordgo.ChatApplicationCommand,
	//Options:     []*discordgo.ApplicationCommandOption{},
}

// ConfigureHandler will take a command request from Discord and translate
// that into an action. This is where we decide if we're taking action
// or if Discord is just asking what autocomplete options are available
func ConfigureHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println(i.Type)
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		configureCommand(s, i)
	default:
		log.Printf("Could not handle %s\n", i.Type)
	}
}

// Actually do the command the user is requesting
func configureCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	allSkills, err := hiscores.GetAllSkills()
	if err != nil {
		log.Println(err)
		return
	}

	var existingConfig model.Servers
	existingConfig, err = storage.FetchServer(i.GuildID)
	if err != nil {
		log.Printf("Unable to fetch an existing config for guild %s. Error: %s", i.GuildID, err)
	}

	existingMessages, err := storage.FetchAllMessages(i.GuildID)
	if err != nil {
		log.Printf("Unable to fetch existing tracked activities for guild %s. Error: %s", i.GuildID, err)
	}

	existingActivities := []string{}
	for _, m := range existingMessages {
		existingActivities = append(existingActivities, m.Activity)
	}

	defaultChannel := []discordgo.SelectMenuDefaultValue{}
	channel, err := utils.GetChannel(s, i.GuildID, existingConfig.ChannelName)
	if err != nil {
		log.Printf(
			"Unable to get channel details for %s in guild %s. Error: %s",
			existingConfig.ChannelName,
			i.GuildID,
			err,
		)
	} else {
		defaultChannel = append(
			defaultChannel,
			discordgo.SelectMenuDefaultValue{
				ID:   channel.ID,
				Type: discordgo.SelectMenuDefaultValueChannel,
			},
		)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modals_survey_configure_" + i.Interaction.Member.User.ID,
			Title:    "Modals survey",
			Components: []discordgo.MessageComponent{
				discordgo.Label{
					Label: "Which channel do you want hiscores posted to?",
					Component: discordgo.SelectMenu{
						MenuType:      discordgo.ChannelSelectMenu,
						CustomID:      "channel",
						Placeholder:   "Choose the Text Channel where you'd like hiscores to be posted",
						DefaultValues: defaultChannel,
						ChannelTypes:  []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "schedule",
							Label:       "Cron Schedule to Update Hiscores (CST)",
							Placeholder: "0 19 * * SUN",
							Style:       discordgo.TextInputShort,
							Required:    true,
							Value:       existingConfig.Schedule,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "activities",
							Label:       "Activites and skills to track",
							Style:       discordgo.TextInputParagraph,
							Placeholder: fmt.Sprint(strings.Join(allSkills[0:6], ",")),
							Required:    true,
							Value:       strings.Join(existingActivities, ","),
							MaxLength:   2000,
						},
					},
				},
				discordgo.Label{
					Label: "re-use existing message for updates?",
					Component: discordgo.SelectMenu{
						MenuType:    discordgo.StringSelectMenu,
						CustomID:    "edit",
						Placeholder: "Edit message instead of posting new?",
						Options: []discordgo.SelectMenuOption{
							{Label: "Yes", Value: "true", Default: existingConfig.ShouldEditMessage == true},
							{Label: "No", Value: "false", Default: existingConfig.ShouldEditMessage == false},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
}

// ConfigureModalSubmit takes action when the users presses submit on the modal survey
// used to configure the server's settings including scheduling, tracked users, and activities
func ConfigureModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Defer our message so we have time to do processing
	// before discord times us out (we get 15 minutes now)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Validating response data. Please wait...",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}

	data := i.ModalSubmitData()

	channelID := data.Components[0].(*discordgo.Label).Component.(*discordgo.SelectMenu).Values[0]
	cronSchedule := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	activities := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	shouldEditMessage, err := strconv.ParseBool(data.Components[3].(*discordgo.Label).Component.(*discordgo.SelectMenu).Values[0])
	if err != nil {
		// It's way less headache to just force a "good" input if the user
		// puts something bogus. This will end up replaced with a menu select later
		// anyways...
		shouldEditMessage = false
	}

	channel, err := s.State.Channel(channelID)

	if err != nil {
		if err == discordgo.ErrStateNotFound {
			log.Printf("Channel %s not found in state cache, fetching from API...", channelID)
			c, err := s.Channel(channelID)
			if err != nil {
				log.Println(err)
				return
			}
			channel = c
		} else {
			log.Println(err)
			return
		}
	}

	// We can't use the state cache because it is
	// populated with all empty values
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Println(err)
		return
	}

	server := model.Servers{
		ID:                guild.ID,
		ServerName:        guild.Name,
		ChannelName:       channel.Name,
		Schedule:          cronSchedule,
		ShouldEditMessage: shouldEditMessage,
		IsEnabled:         true,
	}

	serverErrs := utils.ValidateServerConfig(s, server)
	activityErrs := utils.ValidateActivities(activities)

	if serverErrs != nil || activityErrs != nil {
		errMessage := fmt.Sprintf(
			"\n%s\n%s",
			serverErrs,
			activityErrs,
		)
		cronEmoji := types.ApplicationEmojis["crontab"]
		toolsEmoji := types.ApplicationEmojis["tools"]

		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Flags: discordgo.MessageFlagsEphemeral,

			Content: fmt.Sprintf("Failed to configure channel %s. Reasons are: %s", channel.Name, errMessage),
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
						discordgo.Button{
							Emoji: &discordgo.ComponentEmoji{
								Name: cronEmoji.Name,
								ID:   cronEmoji.ID,
							},
							Label: "Cron Guru",
							Style: discordgo.LinkButton,
							URL:   "https://crontab.guru/#0_19_*_*_SUN",
						},
					},
				},
			},
		})
		if err != nil {
			log.Println(err)
			return
		}

		// Stop any changes from being commited to our DB
		return
	}

	// Once we know what server the user selected we can store that choice
	err = storage.EnrollServer(server, activities)
	if err != nil {
		log.Println(err)
		return
	}

	// Update the cron job for this server in case the user changed the schedule
	schedule.Cron.Remove(schedule.ScheduledJobs[server.ServerName].JobID)
	EnableServerMessageCronjob(server, s)

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Flags: discordgo.MessageFlagsEphemeral,

		Content: fmt.Sprintf("Channel %s is now successfully configured!", channel.Name),
	})

	if err != nil {
		log.Println(err)
		return
	}

	if !strings.HasPrefix(data.CustomID, "modals_survey") {
		return
	}
}
