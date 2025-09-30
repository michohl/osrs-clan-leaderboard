package types

import (
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

// CronSchedule contains all the information required
// to post a message to a discord server on a scheduled
// basis and managed the lifecycle of a cron schedule
type CronSchedule struct {
	JobID          cron.EntryID
	DiscordSession *discordgo.Session
}
