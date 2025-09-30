package discord

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/storage"
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
		panic(err)
	}

	var existingConfig = &storage.ServersRow{}
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
			},
		},
	})
	if err != nil {
		panic(err)
	}
}

// ConfigureModalSubmit takes action when the users presses submit on the modal survey
// used to configure the server's settings including scheduling, tracked users, and activities
func ConfigureModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	fmt.Printf("%+v\n", data)

	channelName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	schedule := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	activities := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		panic(err)
	}

	var channel *discordgo.Channel = &discordgo.Channel{}
	for _, c := range channels {
		if c.Name == channelName {
			channel = c
			break
		}

	}

	if channel.Name == "" {
		panic(fmt.Errorf("Unable to find channel %s in this server", channelName))
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
					panic(err)
				}
				channel = c
			} else {
				panic(err)
			}
		}
	*/

	// We can't use the state cache because it is
	// populated with all empty values
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		panic(err)
	}

	guildID, err := strconv.Atoi(guild.ID)
	if err != nil {
		panic(err)
	}

	// Once we know what server the user selected we can store that choice
	err = storage.EnrollServer(storage.ServersRow{
		ID:          guildID,
		ServerName:  guild.Name,
		ChannelName: channel.Name,
		Activities:  activities,
		Schedule:    schedule,
	})
	if err != nil {
		panic(err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Channel %s is now configured!", channel.Name),
		},
	})
	if err != nil {
		panic(err)
	}

	if !strings.HasPrefix(data.CustomID, "modals_survey") {
		return
	}
}
