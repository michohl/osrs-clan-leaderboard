package utils

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/types"
	"github.com/robfig/cron/v3"
)

// ValidateServerConfig takes a ServersRow struct parsed
// out from our modal survey and then inspects all of the
// individual values to decide if the output is good or not.
//
// If any errors are discovered they'll be returned as an
// "pretty" error that we can send back to the user.
func ValidateServerConfig(s *discordgo.Session, server types.ServersRow) error {

	discoveredErrors := ""

	// TODO: Remove this check once we can use a select menu to only provide
	// the user with valid channels as options
	_, err := GetChannel(s, fmt.Sprintf("%d", server.ID), server.ChannelName)
	if err != nil {
		discoveredErrors = fmt.Sprintf(
			"%s\n* Channel not found in server: %s",
			discoveredErrors,
			server.ChannelName,
		)
	}

	if !IsValidCronExpression(server.Schedule) {
		discoveredErrors = fmt.Sprintf(
			"%s\n* Invalid cron expression: %s",
			discoveredErrors,
			server.Schedule,
		)
	}

	totalActivities := len(strings.Split(server.Activities, ","))
	if totalActivities > 10 {
		discoveredErrors = fmt.Sprintf(
			"%s\n* More than 10 Activities specified: %d",
			discoveredErrors,
			totalActivities,
		)
	}

	for activity := range strings.SplitSeq(server.Activities, ",") {
		_, err := hiscores.IsActivityOrSkill(activity)
		if err != nil {
			discoveredErrors = fmt.Sprintf(
				"%s\n* %s",
				discoveredErrors,
				err,
			)
		}
	}

	if discoveredErrors == "" {
		return nil
	}

	return fmt.Errorf("%s", discoveredErrors)
}

// IsValidCronExpression takes a cron expression and returns whether it is
// a valid expression or not. This is validated by adding it to a dummy cron
func IsValidCronExpression(expression string) bool {
	c := cron.New()
	_, err := c.AddFunc(expression, func() {})
	if err != nil {
		return false
	}
	return true
}
