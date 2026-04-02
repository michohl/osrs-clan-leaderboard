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
func GetUserHiscores(allUsers []model.Users, forceNormalizedLeaderboard bool) (map[model.Users]types.Hiscores, error) {
	var userHiscores map[model.Users]types.Hiscores = make(map[model.Users]types.Hiscores)

	var wg sync.WaitGroup

	// Loading Hiscores for all users in the server
	for _, user := range allUsers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var accountType string
			if forceNormalizedLeaderboard {
				accountType = "main"
			} else {
				accountType = user.OsrsAccountType
			}

			log.Printf("Getting rank for user %s on %s leaderboards\n", user.OsrsUsername, accountType)
			userHS, err := GetPlayerHiscores(user.OsrsUsernameKey, accountType)
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
	leader := -1

	// Prefill the slices unsorted
	for user, hs := range hiscores {
		userRanking := types.RankedUser{
			User:      user,
			LocalRank: -1, // We will adjust this later once we've actually sorted
			Rank:      -1,
			Score:     -1,
			Level:     -1,
			XP:        -1,
		}

		switch activityKind {
		case "activity":
			a := hs.GetActivity(activity)
			if a.Score > leader {
				leader = a.Score
			}
			userRanking.Rank = a.Rank
			userRanking.Score = a.Score
			sortedHiscores.Rankings = append(sortedHiscores.Rankings, userRanking)
		case "skill":
			s := hs.GetSkill(activity)
			if s.Level > leader {
				leader = s.Level
			}
			userRanking.Rank = s.Rank
			userRanking.Level = s.Level
			userRanking.XP = s.XP
			sortedHiscores.Rankings = append(sortedHiscores.Rankings, userRanking)
		}
	}

	// Sort our rankings based on Score or Level
	sort.Slice(sortedHiscores.Rankings, func(i, j int) bool {
		switch activityKind {
		case "activity":
			return sortedHiscores.Rankings[i].Score > sortedHiscores.Rankings[j].Score
		case "skill":
			return sortedHiscores.Rankings[i].Level > sortedHiscores.Rankings[j].Level
		// This switch covers all actual possible values including a default
		// case is necessary to make the compiler happy
		default:
			return false
		}
	})

	// Remove bozos that have never done the content from the list
	if removeUnrankedUsers {
		lastUser := len(sortedHiscores.Rankings) - 1
		for leader > 0 && (sortedHiscores.Rankings[lastUser].Score == 0 || sortedHiscores.Rankings[lastUser].Level == 1) {
			log.Printf(
				"Removing user %s because they have have not done %s\n",
				sortedHiscores.Rankings[lastUser].User.OsrsUsername,
				activity,
			)
			sortedHiscores.Rankings = sortedHiscores.Rankings[:lastUser]
			lastUser = len(sortedHiscores.Rankings) - 1
		}
	}

	// Iterate through sorted list and assign local rankings based on order
	for i := range sortedHiscores.Rankings {
		sortedHiscores.Rankings[i].LocalRank = i + 1
	}

	return &sortedHiscores, nil
}
