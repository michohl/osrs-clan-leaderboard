package hiscores

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Hiscores is the struct we will unmarshall our response from the API into
type Hiscores struct {
	Name       string            `json:"name"`
	Skills     []SkillHiscore    `json:"skills"`
	Activities []ActivityHiscore `json:"activities"`
}

const API_URL = "https://secure.runescape.com/m=hiscore_oldschool/index_lite.json"

// GetPlayerHiscores makes a call to the Jagex provided
// API endpoint that returns the hiscores for one specific user.
// Documentation: https://runescape.wiki/w/Application_programming_interface#Old_School_Hiscores
func GetPlayerHiscores(user OSRSUser) (Hiscores, error) {

	var userHiscores Hiscores

	resp, err := http.Get(
		fmt.Sprintf("%s?player=%s", API_URL, user.EncodeUsername()),
	)
	if err != nil {
		return Hiscores{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Hiscores{}, err
	}

	err = json.Unmarshal(body, &userHiscores)
	if err != nil {
		return Hiscores{}, err
	}

	return userHiscores, nil
}

// GetAllSkills will return all the valid skill options that the API
// can possibly return
func GetAllSkills() ([]string, error) {
	hs, err := GetPlayerHiscores(OSRSUser{Username: "sample"})
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
	hs, err := GetPlayerHiscores(OSRSUser{Username: "sample"})
	if err != nil {
		return []string{}, err
	}

	var discoveredActivities []string
	for _, skill := range hs.Activities {
		discoveredActivities = append(discoveredActivities, skill.Name)
	}

	return discoveredActivities, nil
}
