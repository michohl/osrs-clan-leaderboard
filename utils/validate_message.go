package utils

import (
	"fmt"
	"strings"

	"github.com/michohl/osrs-clan-leaderboard/hiscores"
)

// ValidateActivities takes a csv string of activities and determines if
// all of the activities are valid or not
//
// If any errors are discovered they'll be returned as an
// "pretty" error that we can send back to the user.
func ValidateActivities(activities string) error {

	discoveredErrors := ""

	for activity := range strings.SplitSeq(activities, ",") {
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

	return fmt.Errorf("_Activity Config Issues:_%s", discoveredErrors)
}
