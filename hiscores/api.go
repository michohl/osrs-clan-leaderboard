package hiscores

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/michohl/osrs-clan-leaderboard/types"
)

// HiscoreURLs are the endpoints we reach out to to get Hiscores from Jagex
var HiscoreURLs = map[string]string{
	"ironman":          "https://secure.runescape.com/m=hiscore_oldschool_ironman/index_lite.json",
	"hardcore_ironman": "https://secure.runescape.com/m=hiscore_oldschool_hardcore_ironman/index_lite.json",
	"ultimate_ironman": "https://secure.runescape.com/m=hiscore_oldschool_ultimate/index_lite.json",
	"main":             "https://secure.runescape.com/m=hiscore_oldschool/index_lite.json",
}

// EncodeRSN takes the "human friendly" version
// of the OSRS Username and prepares it for use in
// our API call
func EncodeRSN(username string) string {
	return strings.ToLower(strings.ReplaceAll(username, " ", "_"))
}

// GetPlayerHiscores makes a call to the Jagex provided
// API endpoint that returns the hiscores for one specific user.
// Documentation: https://runescape.wiki/w/Application_programming_interface#Old_School_Hiscores
func GetPlayerHiscores(username string) (types.Hiscores, error) {

	encodedUsername := EncodeRSN(username)

	var userHiscores types.Hiscores

	// We always want to use the main leaderboards for our local rankings
	resp, err := http.Get(
		fmt.Sprintf("%s?player=%s", HiscoreURLs["main"], encodedUsername),
	)
	if err != nil {
		return types.Hiscores{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Hiscores{}, err
	}

	err = json.Unmarshal(body, &userHiscores)
	if err != nil {
		return types.Hiscores{}, err
	}

	return userHiscores, nil
}

// GetAllSkills will return all the valid skill options that the API
// can possibly return
func GetAllSkills() ([]string, error) {
	hs, err := GetPlayerHiscores("sample")
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
	hs, err := GetPlayerHiscores("sample")
	if err != nil {
		return []string{}, err
	}

	var discoveredActivities []string
	for _, skill := range hs.Activities {
		discoveredActivities = append(discoveredActivities, skill.Name)
	}

	return discoveredActivities, nil
}
