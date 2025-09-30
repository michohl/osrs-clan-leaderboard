// Package main is the entrypoint for our application
package main

import (
	"github.com/michohl/osrs-clan-leaderboard/discord"
	"github.com/michohl/osrs-clan-leaderboard/schedule"
)

func main() {
	schedule.Cron.Start()

	// Listen for requests from Discord
	discord.StartBotListener()
}
