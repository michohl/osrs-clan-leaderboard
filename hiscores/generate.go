package hiscores

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/types"

	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
)

// FormatEmbeds takes an activity and user hiscores and formats that information into our final
// set of embeds that we'll pass back to discord to present to the user in the message
func FormatEmbeds(activity string, userHiscores map[model.Users]types.Hiscores, removeUnrankedUsers bool) (*discordgo.MessageEmbed, error) {
	activity = strings.Trim(activity, " ")
	log.Printf("Generating fields for Activity/Skill %s\n", activity)

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

	sortedUserHiscores, err := SortHiscores(userHiscores, activity, removeUnrankedUsers)
	if err != nil {
		return nil, err
	}

	if len(sortedUserHiscores.Rankings) == 0 {
		return nil, nil
	}

	for _, rankedUser := range sortedUserHiscores.Rankings {

		userField.Value = fmt.Sprintf("%s\n", userField.Value)

		if len(sortedUserHiscores.Rankings) > 1 {
			userField.Value += fmt.Sprintf(" %d -", rankedUser.LocalRank)
		}

		userField.Value += fmt.Sprintf(" %s", rankedUser.User.OsrsUsername)

		accountTypeEmoji, ok := types.ApplicationEmojis[rankedUser.User.OsrsAccountType]
		if ok {
			userField.Value += fmt.Sprintf(" <:%s>", accountTypeEmoji.APIName())
		}

		if rankedUser.User.DiscordUserID != "" {
			userField.Value += fmt.Sprintf(" <@%s>", rankedUser.User.DiscordUserID)
		}

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

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s %s", emoji, activity),
		Fields: []*discordgo.MessageEmbedField{
			&userField,
			&quantifierField,
			&rankField,
		},
	}, nil
}

// GetUserHiscores takes a list of users and returns a map populated with all of the
// hiscores for each user
func GetUserHiscores(allUsers []model.Users) (map[model.Users]types.Hiscores, error) {
	var userHiscores map[model.Users]types.Hiscores = make(map[model.Users]types.Hiscores)

	var wg sync.WaitGroup

	// Loading Hiscores for all users in the server
	for _, user := range allUsers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("Getting rank for user %s\n", user.OsrsUsername)
			userHS, err := GetPlayerHiscores(user.OsrsUsernameKey)
			if err != nil {
				// If a user changes their RSN we don't want to break the entire process.
				// We'll just exclude them from the results.
				return
			}

			userHiscores[user] = userHS
		}()
	}

	wg.Wait()

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
func SortHiscores(hiscores map[model.Users]types.Hiscores, activity string, removeUnrankedUsers bool) (*types.RankedHiscores, error) {

	sortedHiscores := types.RankedHiscores{
		Activity: activity,
		Rankings: []types.RankedUser{},
	}

	activityKind, err := IsActivityOrSkill(activity)
	if err != nil {
		return nil, err
	}

	// This isn't used for sorting but we just need some simple way
	// to keep track if _any_ of the users are ranked to avoid an infinite loop later
	rankLeader := -1

	// Prefill the slices unsorted
	for user, hs := range hiscores {
		switch activityKind {
		case "activity":
			a := hs.GetActivity(activity)
			if a.Rank > rankLeader {
				rankLeader = a.Rank
			}
			sortedHiscores.Rankings = append(
				sortedHiscores.Rankings,
				types.RankedUser{
					User:      user,
					LocalRank: -1, // We will adjust this later once we've actually sorted
					Rank:      a.Rank,
					Score:     a.Score,
				},
			)
		case "skill":
			s := hs.GetSkill(activity)
			if s.Rank > rankLeader {
				rankLeader = s.Rank
			}
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

	// Sort our rankings based on the offical rank
	sort.Slice(sortedHiscores.Rankings, func(i, j int) bool {
		return sortedHiscores.Rankings[i].Rank < sortedHiscores.Rankings[j].Rank
	})

	// Remove unranked bozos from the list
	if removeUnrankedUsers {
		for rankLeader > -1 && sortedHiscores.Rankings[0].Rank == -1 {
			sortedHiscores.Rankings = sortedHiscores.Rankings[1:]
		}
	}

	// Iterate through sorted list and assign local rankings based on order
	for i := range sortedHiscores.Rankings {
		sortedHiscores.Rankings[i].LocalRank = i + 1
	}

	return &sortedHiscores, nil
}
