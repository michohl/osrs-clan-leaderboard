package utils

import (
	"log"
	"slices"

	"github.com/michohl/osrs-clan-leaderboard/hiscores"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/storage"
)

func RefreshAccountTypes(users []model.Users) []model.Users {

	refreshedUsers := []model.Users{}

	// Not all game modes have leaderboards so it's not possible to truly discover the correct account type every time.
	// If the user enrolls to the server and specifies one of these game modes we'll just never refresh their account type
	// in the database so we can accurately represent their icon in the rendered messages
	excludedAccountTypes := []string{
		"main",
		"group_ironman",
		"hardcore_group_ironman",
		"unranked_group_ironman",
	}

	for _, user := range users {
		// If the user already claims to be on the main leaderboard then there's no point in checking the others.
		// If the user claims to be on a specialized leaderboard then we have to make sure their status hasn't changed.
		// e.g. Hardcore dies and becomes regular Iron or Iron de-irons and becomes a main
		if user.OsrsAccountType == "" || !slices.Contains(excludedAccountTypes, user.OsrsAccountType) {
			discoveredAccountType := hiscores.DiscoverRSAccountType(user.OsrsUsername)
			if user.OsrsAccountType != discoveredAccountType {
				log.Printf("User %s has changed account type from %s to %s. Updating DB to reflect this change\n", user.OsrsUsername, user.OsrsAccountType, discoveredAccountType)
				user.OsrsAccountType = discoveredAccountType
				err := storage.EnrollUser(user)
				if err != nil {
					log.Printf("Failed to update account type for  user %s\n", user.OsrsUsername)
				}
			}
		}

		refreshedUsers = append(refreshedUsers, user)
	}

	return refreshedUsers
}
