package hiscores

import (
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/types"

	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
)

func getEmbedSize(embed *discordgo.MessageEmbed) int {
	size := 0
	for _, f := range embed.Fields {
		size += len(f.Value)
	}

	return size
}

func IsSeasonal(activity string) bool {
	seasonalMarkers := []string{"league", "leagues", "deadman", "deadman mode"}
	for _, sm := range seasonalMarkers {
		suffix := fmt.Sprintf("(%s)", sm)
		if strings.HasSuffix(strings.ToLower(activity), suffix) {
			return true
		}
	}

	return false
}

// FormatEmbeds takes an activity and user hiscores and formats that information into our final
// set of embeds that we'll pass back to discord to present to the user in the message
func FormatEmbeds(activity string, userHiscores map[model.Users]types.Hiscores, removeUnrankedUsers bool, removeRank bool) ([]*discordgo.MessageEmbed, error) {
	isSeasonalUsers := true
	for user := range userHiscores {
		if user.OsrsAccountType != "seasonal" {
			isSeasonalUsers = false
			break
		}
	}

	emojiName := types.NormalizeEmojiName(activity)
	activityEmoji, ok := types.ApplicationEmojis[emojiName]
	if !ok {
		activityEmoji = types.ApplicationEmojis["osrstrophy"]
	}

	emoji := fmt.Sprintf("<:%s>", activityEmoji.APIName())

	if isSeasonalUsers && !slices.Contains(types.SEASONAL_ACTIVITIES, strings.ToLower(activity)) {
		seasonalEmoji := types.ApplicationEmojis["league_points"]
		emoji += fmt.Sprintf(" <:%s>", seasonalEmoji.APIName())
	}

	activity = strings.Trim(activity, " ")
	log.Printf("Generating fields for Activity/Skill %s\n", activity)

	activityKind, err := IsActivityOrSkill(activity)
	if err != nil {
		return nil, err
	}

	var quantifierHeader string
	switch activityKind {
	case "activity":
		quantifierHeader = "Score"
	case "skill":
		quantifierHeader = "Level"
	}

	messageEmbeds := []*discordgo.MessageEmbed{
		{
			Title: fmt.Sprintf("%s %s", emoji, activity),
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Username", Value: "", Inline: true},       // User field
				{Name: quantifierHeader, Value: "", Inline: true}, // Quantifier field
				{Name: "Rank", Value: "", Inline: true},           // Rank field
			},
		},
	}

	currentEmbedIndex := len(messageEmbeds) - 1
	currentEmbed := messageEmbeds[currentEmbedIndex]

	userField := currentEmbed.Fields[0]
	quantifierField := currentEmbed.Fields[1]
	rankField := currentEmbed.Fields[2]

	var cleanActivity string
	if IsSeasonal(activity) && strings.LastIndex(activity, "(") > -1 {
		cleanActivity = activity[:strings.LastIndex(activity, "(")]
	} else {
		cleanActivity = activity
	}
	sortedUserHiscores, err := SortHiscores(userHiscores, cleanActivity, removeUnrankedUsers)
	if err != nil {
		return nil, err
	}

	if len(sortedUserHiscores.Rankings) == 0 {
		return nil, nil
	}

	for _, rankedUser := range sortedUserHiscores.Rankings {

		// If we get close to the character limit (1024) then we should split
		// the embed into multiple messages
		if getEmbedSize(currentEmbed) > 850 {
			messageEmbeds = append(messageEmbeds, &discordgo.MessageEmbed{
				Title: fmt.Sprintf("%s %s", emoji, activity),
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Username", Value: "", Inline: true},       // User field
					{Name: quantifierHeader, Value: "", Inline: true}, // Quantifier field
					{Name: "Rank", Value: "", Inline: true},           // Rank field
				},
			},
			)

			currentEmbedIndex = len(messageEmbeds) - 1
			currentEmbed = messageEmbeds[currentEmbedIndex]

			userField = currentEmbed.Fields[0]
			quantifierField = currentEmbed.Fields[1]
			rankField = currentEmbed.Fields[2]
		}

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

	if removeRank {
		for _, embed := range messageEmbeds {
			embed.Fields = embed.Fields[:2]
		}
	}

	return messageEmbeds, nil
}

// GetUserHiscores takes a list of users and returns a map populated with all of the
// hiscores for each user
func GetUserHiscores(allUsers []model.Users, leaderboardOverride string) (map[model.Users]types.Hiscores, error) {
	var userHiscores map[model.Users]types.Hiscores = make(map[model.Users]types.Hiscores)

	// Loading Hiscores for all users in the server
	for _, user := range allUsers {
		if leaderboardOverride != "" {
			user.OsrsAccountType = leaderboardOverride
		}

		log.Printf("Getting rank for user %s on %s leaderboards\n", user.OsrsUsername, user.OsrsAccountType)
		userHS, err := GetPlayerHiscores(user)

		// If a user changes their RSN we don't want to break the entire process.
		// We'll just exclude them from the results.
		if err == nil {
			userHiscores[user] = userHS
		}
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
func SortHiscores(hiscores map[model.Users]types.Hiscores, activity string, removeUnrankedUsers bool) (*types.RankedHiscores, error) {

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

			userRanking.Rank = a.Rank
			userRanking.Score = a.Score

			// For some reason some users who haven't done content have a score of
			// -1 while others have a score of 0 so we're just going to normalize
			// that and pretend we didn't see that weirdness
			if userRanking.Score == -1 {
				userRanking.Score = 0
			}
		case "skill":
			s := hs.GetSkill(activity)

			userRanking.Rank = s.Rank
			userRanking.Level = s.Level
			userRanking.XP = s.XP

			if userRanking.Level == -1 {
				userRanking.Level = 1
			}
		}

		sortedHiscores.Rankings = append(sortedHiscores.Rankings, userRanking)
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
		// Get the last place user
		lastUserIndex := len(sortedHiscores.Rankings) - 1
		if lastUserIndex == -1 {
			return &sortedHiscores, nil
		}
		lastUser := sortedHiscores.Rankings[lastUserIndex]

		// Keep removing the last place user until the last place user
		// has done the content enough that's worth representing
		for lastUser.Score == 0 || lastUser.Level == 1 {
			log.Printf(
				"Removing user %s because they have have not done %s\n",
				lastUser.User.OsrsUsername,
				activity,
			)
			sortedHiscores.Rankings = sortedHiscores.Rankings[:lastUserIndex]

			// Refresh who the last place user is
			lastUserIndex = len(sortedHiscores.Rankings) - 1
			if lastUserIndex == -1 {
				log.Printf("All users have been filtered out for %s %s", activityKind, activity)
				return &sortedHiscores, nil
			}

			lastUser = sortedHiscores.Rankings[lastUserIndex]
		}
	}

	// Iterate through sorted list and assign local rankings based on order
	for i := range sortedHiscores.Rankings {
		sortedHiscores.Rankings[i].LocalRank = i + 1
	}

	return &sortedHiscores, nil
}
