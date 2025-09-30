package schedule

import (
	"github.com/michohl/osrs-clan-leaderboard/types"
	"github.com/robfig/cron/v3"
)

var (
	// Cron is the global cron that runs in
	// our Go process and enrolled servers use
	// for their scheduled hiscore updates
	Cron *cron.Cron

	// ScheduledJobs is how we keep track of all of our scheduled jobs
	// in our running process so we can manage their lifecycle
	ScheduledJobs map[string]types.CronSchedule = make(map[string]types.CronSchedule)
)

func init() {
	Cron = cron.New()

}
