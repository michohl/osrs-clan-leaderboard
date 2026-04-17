package hiscores

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// HiscoreModes is the `m=` parameter on our API URI we reach out to to get Hiscores from Jagex
var HiscoreModes = map[string]string{
	"hardcore_ironman": "hiscore_oldschool_hardcore_ironman",
	"ultimate_ironman": "hiscore_oldschool_ultimate",
	"ironman":          "hiscore_oldschool_ironman",
	"main":             "hiscore_oldschool",

	// Special game modes like leagues or deadman mode
	"seasonal": "hiscore_oldschool_seasonal",

	// These modes don't have their own leaderboard from the API so we just have to use the main
	// "unranked_group_ironman": "hiscore_oldschool",
	// "group_ironman":          "hiscore_oldschool",
	// "hardcore_group_ironman": "hiscore_oldschool",
}

// EncodeRSN takes the "human friendly" version
// of the OSRS Username and prepares it for use in
// our API call
func EncodeRSN(username string) string {
	return strings.ToLower(strings.ReplaceAll(username, " ", "_"))
}

func queryAPI(user model.Users) (types.Hiscores, error) {
	encodedUsername := EncodeRSN(user.OsrsUsername)

	var userHiscores types.Hiscores

	mode, ok := HiscoreModes[user.OsrsAccountType]
	if !ok {
		log.Printf("Account Type '%s' does not have an associated Hiscores mode. Defaulting to Overall hiscores\n", user.OsrsAccountType)
		mode = HiscoreModes["main"]
	}

	resp, err := http.Get(
		fmt.Sprintf("https://secure.runescape.com/m=%s/index_lite.json?player=%s", mode, encodedUsername),
	)

	if err != nil {
		log.Printf(
			"Unable to fetch hiscore for user %s (%s) on %s leaderboard\n%s\n",
			user.OsrsUsername,
			encodedUsername,
			user.OsrsAccountType,
			err,
		)

		return types.Hiscores{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read body from HTTP response")
		return types.Hiscores{}, err
	}

	err = json.Unmarshal(body, &userHiscores)
	if err != nil {
		log.Printf("Unable to unmarshal body as JSON: %s\n", err)
		return types.Hiscores{}, fmt.Errorf("No valid hiscores returned for user %s on %s leaderboard", user.OsrsUsername, user.OsrsAccountType)
	}

	return userHiscores, nil

}

// GetPlayerHiscores makes a call to the Jagex provided
// API endpoint that returns the hiscores for one specific user.
// Documentation: https://runescape.wiki/w/Application_programming_interface#Old_School_Hiscores
func GetPlayerHiscores(user model.Users) (types.Hiscores, error) {
	hiscores, err := retry.NewWithData[types.Hiscores](retry.Attempts(5), retry.Delay(100*time.Millisecond)).Do(
		func() (types.Hiscores, error) {
			hiscores, err := queryAPI(user)
			if err != nil {
				return types.Hiscores{}, err
			}
			return hiscores, nil
		},
	)

	if err != nil {
		return types.Hiscores{}, err
	}

	return hiscores, nil
}

// GetAllSkills will return all the valid skill options that the API
// can possibly return
func GetAllSkills() ([]string, error) {
	hs, err := GetPlayerHiscores(model.Users{OsrsUsername: "sample", OsrsAccountType: "main"})
	if err != nil {
		return []string{}, err
	}

	var discoveredSkills []string
	for _, skill := range hs.Skills {
		discoveredSkills = append(discoveredSkills, skill.Name)
	}

	return discoveredSkills, nil
}

// GetAllActivities will return all the valid skill options that the API
// can possibly return
func GetAllActivities() ([]string, error) {
	hs, err := GetPlayerHiscores(model.Users{OsrsUsername: "sample", OsrsAccountType: "main"})
	if err != nil {
		return []string{}, err
	}

	var discoveredActivities []string
	for _, skill := range hs.Activities {
		discoveredActivities = append(discoveredActivities, skill.Name)
	}

	return discoveredActivities, nil
}

// GuessUserAccountType takes a RSN and checks all the available
// leaderboards to see if we can determine what kind of account
// the user actually is. Default to main if nothing more suitable found
func GuessUserAccountType(username string) string {
	encodedUsername := EncodeRSN(username)

	accountTypes := sortAccountTypes()

	for _, accountType := range accountTypes {
		mode := HiscoreModes[accountType]
		resp, _ := http.Get(
			fmt.Sprintf("https://secure.runescape.com/m=%s/index_lite.json?player=%s", mode, encodedUsername),
		)

		if resp.StatusCode != 404 {
			return accountType
		}
	}

	return "main"
}

// sortAccountTypes takes the HiscoreModes and sorts the keys
// to prioritize the specialized game modes and then end with
// ironman then main.
func sortAccountTypes() []string {

	lowPriorityAccountTypes := []string{"ironman", "main"}
	accountTypes := []string{}

	for accountType := range HiscoreModes {
		if !slices.Contains(lowPriorityAccountTypes, accountType) {
			accountTypes = append(accountTypes, accountType)
		}
	}

	slices.Sort(accountTypes)
	return append(accountTypes, lowPriorityAccountTypes...)

}
