package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"

	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
)

// ValidateServerConfig takes a ServersRow struct parsed
// out from our modal survey and then inspects all of the
// individual values to decide if the output is good or not.
//
// If any errors are discovered they'll be returned as an
// "pretty" error that we can send back to the user.
func ValidateServerConfig(s *discordgo.Session, server model.Servers) error {

	discoveredErrors := ""

	if !IsValidCronExpression(server.Schedule) {
		discoveredErrors = fmt.Sprintf(
			"%s\n* Invalid cron expression: %s",
			discoveredErrors,
			server.Schedule,
		)
	}

	if discoveredErrors == "" {
		return nil
	}

	return fmt.Errorf("_Server Config Issues:_%s", discoveredErrors)
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
