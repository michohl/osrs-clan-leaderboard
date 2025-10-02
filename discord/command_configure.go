package discord

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
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
		log.Fatalf("Could not handle %s", i.Type)
	}
}

// Actually do the command the user is requesting
func configureCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	allSkills, err := hiscores.GetAllSkills()
	if err != nil {
		log.Fatal(err)
		return
	}

	var existingConfig = &types.ServersRow{}
	existingConfig, err = storage.FetchServer(i.GuildID)
	if err != nil {
		log.Printf("Unable to fetch an existing config for guild %s. Error: %s", i.GuildID, err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modals_survey_configure_" + i.Interaction.Member.User.ID,
			Title:    "Modals survey",
			Components: []discordgo.MessageComponent{
				/*
					// This won't work until https://github.com/bwmarrin/discordgo/pull/1656
					// is merged
					discordgo.Label{
						Label: "Which channel do you want hiscores posted to?",
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								MenuType:     discordgo.ChannelSelectMenu,
								CustomID:     "channel",
								Placeholder:  "Choose the Text Channel where you'd like hiscores to be posted",
								ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
							},
						},
					},
				*/
				// TODO: Replace this with a Label + SelectMenu once Label support
				// is merged into the discordgo library
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "channel",
							Label:       "Which channel to send messages to?",
							Placeholder: "#something",
							Style:       discordgo.TextInputShort,
							Required:    true,
							Value:       existingConfig.ChannelName,
						},
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
							Value:       existingConfig.Activities,
							MaxLength:   2000,
						},
					},
				},
				// TODO: Replace with a dropdown menu with yes/no options
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "edit",
							Label:       "Edit message instead of posting new?",
							Style:       discordgo.TextInputShort,
							Placeholder: "true",
							Required:    true,
							Value:       fmt.Sprintf("%t", existingConfig.ShouldEditMessage),
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
		log.Fatal(err)
		return
	}

	data := i.ModalSubmitData()

	channelName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	cronSchedule := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	activities := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	shouldEditMessage, err := strconv.ParseBool(data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value)
	if err != nil {
		// It's way less headache to just force a "good" input if the user
		// puts something bogus. This will end up replaced with a menu select later
		// anyways...
		shouldEditMessage = false
	}

	// TODO: Switch back to this method once we get access to
	// Labels with SelectMenus to get the Channel SelectMenu
	/*
				channel, err := s.State.Channel(channelID)

				if err != nil {
					if err == discordgo.ErrStateNotFound {
						log.Printf("Channel %s not found in state cache, fetching from API...", channelID)
						c, err := s.Channel(channelID)
						if err != nil {
							log.Fatal(err)
		        return
						}
						channel = c
					} else {
						log.Fatal(err)
		        return
					}
				}
	*/

	// We can't use the state cache because it is
	// populated with all empty values
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Fatal(err)
		return
	}

	guildID, err := strconv.Atoi(guild.ID)
	if err != nil {
		log.Fatal(err)
		return
	}

	server := types.ServersRow{
		ID:                guildID,
		ServerName:        guild.Name,
		ChannelName:       channelName,
		Activities:        activities,
		Schedule:          cronSchedule,
		ShouldEditMessage: shouldEditMessage,
	}

	err = utils.ValidateServerConfig(s, server)
	if err != nil {
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Flags: discordgo.MessageFlagsEphemeral,

			Content: fmt.Sprintf("Failed to configure channel %s. Reasons are: %s", channelName, err),
		})
		if err != nil {
			log.Fatal(err)
			return
		}

		// Stop any changes from being commited to our DB
		return
	}

	// Once we know what server the user selected we can store that choice
	err = storage.EnrollServer(server)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Update the cron job for this server in case the user changed the schedule
	schedule.Cron.Remove(schedule.ScheduledJobs[server.ServerName].JobID)
	EnableServerMessageCronjob(&server, s)

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Flags: discordgo.MessageFlagsEphemeral,

		Content: fmt.Sprintf("Channel %s is now successfully configured!", channelName),
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	if !strings.HasPrefix(data.CustomID, "modals_survey") {
		return
	}
}
