package discord

import (
	"log"
	"slices"
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

	userHiscores, err := hiscores.GetUserHiscores(allUsers)
	if err != nil {
		return err
	}

	log.Printf("Generating %d Hiscores messages for server %s", len(messages), server.ServerName)

	var wg sync.WaitGroup

	// Keep track of each activity message that has been posted
	// so we can prevent threads from skipping past each other
	positionsPosted := []int{-1}

	for i, activityMessage := range messages {
		wg.Add(1)
		go func() error {
			defer wg.Done()
			log.Printf("Generating Hiscores message for activity %s", activityMessage.Activity)
			he, err := hiscores.FormatEmbeds(activityMessage.Activity, userHiscores)
			if err != nil {
				return err
			}

			// Wait until the previous message in the sequence is posted
			for !slices.Contains(positionsPosted, i-1) {
				continue
			}

			// We remove the message every time instead of editing the existing message so we can guarantee
			// that the order the skills are posted in matches the order the server admin configured
			if activityMessage.MessageID != "" && server.ShouldEditMessage {
				log.Printf("Removing existing hiscores message for %s in server %s\n", activityMessage.Activity, server.ServerName)
				err := s.ChannelMessageDelete(channel.ID, activityMessage.MessageID)
				if err != nil {
					return err
				}

				err = storage.RemoveMessage(activityMessage)
				if err != nil {
					return err
				}
			}

			log.Printf("Posting new scheduled hiscores message for %s in server %s\n", activityMessage.Activity, server.ServerName)
			newMessage, err := s.ChannelMessageSendEmbed(channel.ID, he)
			if err != nil {
				return err
			}

			positionsPosted = append(positionsPosted, i)

			activityMessage.MessageID = newMessage.ID

			// Update our records to keep track of this message so we can edit it later
			err = storage.EnrollMessage(server, activityMessage)
			if err != nil {
				return err
			}

			return nil

		}()
	}

	wg.Wait()

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
