package discord

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/schedule"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// PostHiscoresMessage posts a message to the user configured
// channel. This is done on a cron configured by the user
func PostHiscoresMessage(server *types.ServersRow, s *discordgo.Session) error {
	log.Printf("Posting scheduled hiscores message for %s\n", server.ServerName)

	channel, err := GetChannel(s, fmt.Sprintf("%d", server.ID), server.ChannelName)
	if err != nil {
		return err
	}

	s.ChannelMessageSend(channel.ID, "Hello from the cron!")

	return nil
}

// EnableServerMessageCronjob takes information about all of our
// enrolled servers and starts a cronjob to post their hiscores
// update messages on the configured schedule
//
// We moved this outside the schedule package to avoid some circular imports
func EnableServerMessageCronjob(server *types.ServersRow, s *discordgo.Session) error {

	jobID, err := schedule.Cron.AddFunc(server.Schedule, func() {
		PostHiscoresMessage(server, s)
	})
	if err != nil {
		log.Fatalf("Unable to schedule cron job for server %s because %s", server.ServerName, err)
	}

	log.Printf("Cron successfully scheduled for server %s. Job ID %d\n", server.ServerName, jobID)
	schedule.ScheduledJobs[server.ServerName] = types.CronSchedule{
		JobID:          jobID,
		DiscordSession: s,
	}

	return nil
}
