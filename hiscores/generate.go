package hiscores

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// GenerateHiscoresFields takes a specific server and finds all the
// activities that server is tracking and generates hiscores with the
// users enrolled in that specific server
func GenerateHiscoresFields(server *types.ServersRow) ([]*discordgo.MessageEmbed, error) {
	log.Printf("Generating Hiscores for server %s", server.ServerName)

	allActivities := strings.Split(server.Activities, ",")
	allUsers, err := storage.FetchAllUsers(fmt.Sprintf("%d", server.ID))
	if err != nil {
		return nil, err
	}

	userHiscores, err := GetUserHiscores(allUsers)
	if err != nil {
		return nil, err
	}

	var embeds []*discordgo.MessageEmbed

	for _, activity := range allActivities {
		activity = strings.Trim(activity, " ")
		log.Printf("Generating fields for Acitivity/Skill %s\n", activity)

		activityKind, err := IsActivityOrSkill(activity)
		if err != nil {
			return nil, err
		}

		userField := discordgo.MessageEmbedField{
			Name:   "Username",
			Value:  "",
			Inline: true,
		}

		quantifierField := discordgo.MessageEmbedField{
			Name:   "",
			Value:  "",
			Inline: true,
		}

		rankField := discordgo.MessageEmbedField{
			Name:   "Rank",
			Value:  "",
			Inline: true,
		}

		switch activityKind {
		case "activity":
			quantifierField.Name = "Score"
		case "skill":
			quantifierField.Name = "Level"
		}

		sortedUserHiscores, err := SortHiscores(userHiscores, activity)
		if err != nil {
			return nil, err
		}

		for _, rankedUser := range sortedUserHiscores.Rankings {

			userField.Value = fmt.Sprintf(
				"%s\n%d - %s <@%d>",
				userField.Value,
				rankedUser.LocalRank,
				rankedUser.User.OSRSUsername,
				rankedUser.User.DiscordUserID,
			)

			switch activityKind {
			case "skill":
				quantifierField.Value = fmt.Sprintf(
					"%s\n%d",
					quantifierField.Value,
					rankedUser.Level,
				)
			case "activity":
				quantifierField.Value = fmt.Sprintf(
					"%s\n%d",
					quantifierField.Value,
					rankedUser.Score,
				)
			}

			rankField.Value = fmt.Sprintf(
				"%s\n%d",
				rankField.Value,
				rankedUser.Rank,
			)

		}

		emojiName := types.NormalizeEmojiName(activity)

		activityEmoji, ok := types.ApplicationEmojis[emojiName]
		if !ok {
			activityEmoji = types.ApplicationEmojis["osrstrophy"]
		}

		emoji := fmt.Sprintf("<:%s>", activityEmoji.APIName())

		embeds = append(embeds, &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s %s", emoji, activity),
			Fields: []*discordgo.MessageEmbedField{
				&userField,
				&quantifierField,
				&rankField,
			},
		})

	}

	return embeds, nil
}

// GetUserHiscores takes a list of users and returns a map populated with all of the
// hiscores for each user
func GetUserHiscores(allUsers []*types.UsersRow) (map[types.UsersRow]types.Hiscores, error) {
	var userHiscores map[types.UsersRow]types.Hiscores = make(map[types.UsersRow]types.Hiscores)

	// Loading Hiscores for all users in the server
	for _, user := range allUsers {
		log.Printf("Getting rank for user %s\n", user.OSRSUsername)
		userHS, err := GetPlayerHiscores(types.OSRSUser{Username: user.OSRSUsername})
		if err != nil {
			return nil, err
		}

		userHiscores[*user] = userHS
	}

	return userHiscores, nil
}

// ContainsCaseInsensitive checks if a string slice contains a specific string, ignoring case.
func ContainsCaseInsensitive(slice []string, target string) bool {
	for _, element := range slice {
		if strings.EqualFold(element, target) {
			return true
		}
	}
	return false
}

// IsActivityOrSkill takes a user specified activity or skill
// and compares it to all of the known activities and skills to determine
// what the "kind" is. Either skill or activity
func IsActivityOrSkill(name string) (string, error) {

	name = strings.Trim(name, " ")

	allActivities, err := GetAllActivities()
	if err != nil {
		return "", err
	}

	allSkills, err := GetAllSkills()
	if err != nil {
		return "", err
	}

	switch {
	case ContainsCaseInsensitive(allActivities, name):
		return "activity", nil
	case ContainsCaseInsensitive(allSkills, name):
		return "skill", nil
	default:
		return "", fmt.Errorf("Unable to associate %s with any known skill or activity", name)
	}

}

// SortHiscores takes an array of hiscores for multiple
// users and sorts them by rank for a specific activity
func SortHiscores(hiscores map[types.UsersRow]types.Hiscores, activity string) (*types.RankedHiscores, error) {

	sortedHiscores := types.RankedHiscores{
		Activity: activity,
		Rankings: []types.RankedUser{},
	}

	activityKind, err := IsActivityOrSkill(activity)
	if err != nil {
		return nil, err
	}

	// Prefill the slices unsorted
	for user, hs := range hiscores {
		switch activityKind {
		case "activity":
			a := hs.GetActivity(activity)
			if a.Rank > 0 {
				sortedHiscores.Rankings = append(
					sortedHiscores.Rankings,
					types.RankedUser{
						User:      user,
						LocalRank: -1, // We will adjust this later once we've actually sorted
						Rank:      a.Rank,
						Score:     a.Score,
					},
				)
			}
		case "skill":
			s := hs.GetSkill(activity)
			if s.Rank > 0 {
				sortedHiscores.Rankings = append(
					sortedHiscores.Rankings,
					types.RankedUser{
						User:      user,
						LocalRank: -1, // We will adjust this later once we've actually sorted
						Rank:      s.Rank,
						Level:     s.Level,
						XP:        s.XP,
					},
				)
			}
		}
	}

	// Sort our rankings based on the offical rank
	sort.Slice(sortedHiscores.Rankings, func(i, j int) bool {
		return sortedHiscores.Rankings[i].Rank < sortedHiscores.Rankings[j].Rank
	})

	// Iterate through sorted list and assign local rankings based on order
	for i := range sortedHiscores.Rankings {
		sortedHiscores.Rankings[i].LocalRank = i + 1
	}

	return &sortedHiscores, nil
}
