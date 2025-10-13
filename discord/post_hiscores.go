package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/schedule"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
	"github.com/michohl/osrs-clan-leaderboard/utils"

	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
)

// PostHiscoresMessage posts a message to the user configured
// channel. This is done on a cron configured by the user
func PostHiscoresMessage(serverID string, s *discordgo.Session) error {

	// Fetch the latest data from our db in case it has changed
	server, err := storage.FetchServer(serverID)
	if err != nil {
		return err
	}

	channel, err := utils.GetChannel(s, server.ID, server.ChannelName)
	if err != nil {
		return err
	}

	hiscoresEmbeds, err := hiscores.GenerateHiscoresFields(server)
	if err != nil {
		return err
	}

	if server.MessageID != "" && server.ShouldEditMessage {
		log.Printf("Editing existing scheduled hiscores message for %s\n", server.ServerName)
		s.ChannelMessageEditEmbeds(channel.ID, server.MessageID, hiscoresEmbeds)
	} else {
		log.Printf("Posting new scheduled hiscores message for %s\n", server.ServerName)
		message, err := s.ChannelMessageSendEmbeds(channel.ID, hiscoresEmbeds)
		if err != nil {
			return err
		}

		// Update our records to keep track of this message so we can edit it later
		err = storage.UpdateServerMessageID(server, message.ID)
		if err != nil {
			return err
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
		PostHiscoresMessage(server.ID, s)
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
