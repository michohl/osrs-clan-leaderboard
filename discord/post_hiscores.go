package discord

import (
	"log"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/schedule"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
	"github.com/michohl/osrs-clan-leaderboard/utils"

	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
)

// PostHiscoresMessages posts a message per activity to the user configured
// channel. This is done on a cron configured by the user
func PostHiscoresMessages(serverID string, s *discordgo.Session) error {

	// Fetch the latest data from our db in case it has changed
	server, err := storage.FetchServer(serverID)
	if err != nil {
		return err
	}

	messages, err := storage.FetchAllMessages(serverID)
	if err != nil {
		return err
	}

	channel, err := utils.GetChannel(s, server.ID, server.ChannelName)
	if err != nil {
		return err
	}

	allUsers, err := storage.FetchAllUsers(server.ID)
	if err != nil {
		return err
	}

	allActivitiesAndSkills, err := storage.FetchAllActivitiesAndSkills(server.ID)
	if err != nil {
		return err
	}

	userHiscores, err := hiscores.GetUserHiscores(allUsers, "")
	if err != nil {
		return err
	}

	userSeasonalHiscores := map[model.Users]types.Hiscores{}
	for _, aos := range allActivitiesAndSkills {
		if hiscores.IsSeasonal(aos) || slices.Contains(types.SEASONAL_ACTIVITIES, strings.ToLower(aos)) {
			log.Println("At least one seasonal activity/skill detected so generating list of seasonal hiscores now...")
			userSeasonalHiscores, err = hiscores.GetUserHiscores(allUsers, "seasonal")
			if err != nil {
				return err
			}
			break
		}
	}

	log.Printf("Generating %d Hiscores messages for server %s", len(messages), server.ServerName)

	// Keep track of each activity message that has been posted
	// so we can prevent threads from skipping past each other
	type preparedHiscoresMessage struct {
		activity string
		embed    *discordgo.MessageEmbed
	}
	preparedEmbeds := map[int][]preparedHiscoresMessage{}

	var wg sync.WaitGroup
	for i, aos := range allActivitiesAndSkills {
		wg.Add(1)
		go func() error {
			defer wg.Done()
			log.Printf("Generating Hiscores message for activity %s", aos)
			var messageEmbeds []*discordgo.MessageEmbed

			if hiscores.IsSeasonal(aos) && strings.LastIndex(aos, "(") != -1 {

			}

			var hs map[model.Users]types.Hiscores

			if hiscores.IsSeasonal(aos) {
				if strings.LastIndex(aos, "(") != -1 {
					aos = aos[:strings.LastIndex(aos, "(")]
				}
				hs = userSeasonalHiscores
			} else if slices.Contains(types.SEASONAL_ACTIVITIES, strings.ToLower(aos)) {
				hs = userSeasonalHiscores
			} else {
				hs = userHiscores
			}

			messageEmbeds, err = hiscores.FormatEmbeds(aos, hs, true, true)

			if err != nil {
				return err
			}

			log.Printf("Generated embeds for %s: %d\n", aos, len(messageEmbeds))

			for _, messageEmbed := range messageEmbeds {
				// If we filter out all of the users from an embed because every user has
				// zero score or level 1 then we can just throw the whole message away
				if messageEmbed != nil {
					preparedEmbeds[i] = append(preparedEmbeds[i], preparedHiscoresMessage{activity: aos, embed: messageEmbed})
				}
			}

			return nil
		}()
	}

	// Wait for all messages to be generated before posting
	wg.Wait()

	log.Println("All Hiscores are generated. Starting to post discord messages")

	for _, key := range slices.Sorted(maps.Keys(preparedEmbeds)) {

		if len(preparedEmbeds[key]) < 1 {
			continue
		}

		activityMessages, err := storage.FetchMessage(serverID, preparedEmbeds[key][0].activity)
		if err != nil {
			return err
		}

		for _, activityMessage := range activityMessages {
			// We remove the message every time instead of editing the existing message so we can guarantee
			// that the order the skills are posted in matches the order the server admin configured
			if activityMessage.MessageID != "" && server.ShouldEditMessage {
				log.Printf("Removing existing hiscores message for %s in server %s\n", activityMessage.Activity, server.ServerName)
				err := s.ChannelMessageDelete(channel.ID, activityMessage.MessageID)
				if err != nil {
					log.Printf("Error removing existing message: %s", err)
				}

				err = storage.RemoveMessage(activityMessage)
				if err != nil {
					log.Printf("Error removing deleted message from db: %s", err)
				}
			}
		}

		for _, embed := range preparedEmbeds[key] {
			he := embed.embed

			log.Printf("Posting new scheduled hiscores message for %s in server %s\n", embed.activity, server.ServerName)
			newMessage, err := s.ChannelMessageSendEmbed(channel.ID, he)
			if err != nil {
				return err
			}

			position, err := storage.FetchActivityPosition(serverID, embed.activity)
			if err != nil {
				return err
			}

			activityMessage := model.Messages{
				MessageID: newMessage.ID,
				ServerID:  serverID,
				Activity:  embed.activity,
				Position:  position,
			}

			// Update our records to keep track of this message so we can edit it later
			err = storage.EnrollMessage(server, activityMessage)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EnableServerMessageCronjob takes information about all of our
// enrolled servers and starts a cronjob to post their hiscores
// update messages on the configured schedule
//
// We moved this outside the schedule package to avoid some circular imports
func EnableServerMessageCronjob(server model.Servers, s *discordgo.Session) error {

	jobID, err := schedule.Cron.AddFunc(server.Schedule, func() {
		PostHiscoresMessages(server.ID, s)
	})
	if err != nil {
		log.Printf("Unable to schedule cron job for server %s because %s\n", server.ServerName, err)
	}

	log.Printf("Cron successfully scheduled for server %s. Job ID %d\n", server.ServerName, jobID)
	schedule.ScheduledJobs[server.ServerName] = types.CronSchedule{
		JobID:          jobID,
		DiscordSession: s,
	}

	return nil
}
