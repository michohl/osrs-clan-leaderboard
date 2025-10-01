package hiscores

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
)

// GenerateHiscoresFields takes a specific server and finds all the
// activities that server is tracking and generates hiscores with the
// users enrolled in that specific server
func GenerateHiscoresFields(server *types.ServersRow) ([]*discordgo.MessageEmbedField, error) {
	log.Printf("Generating Hiscores for server %s", server.ServerName)

	allActivities := strings.Split(server.Activities, ",")
	allUsers, err := storage.FetchAllUsers(fmt.Sprintf("%d", server.ID))
	if err != nil {
		return nil, err
	}

	fmt.Println("Data:")
	fmt.Println(allActivities)
	fmt.Println(allUsers)

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

	var fields []*discordgo.MessageEmbedField

	for _, activity := range allActivities {
		activity = strings.Trim(activity, " ")
		log.Printf("Generating fields for Acitivity/Skill %s\n", activity)

		rank := 1
		fieldValue := ""
		for user, hs := range userHiscores {

			// TODO: Check activities as well
			hiscore := hs.GetSkill(activity)

			fieldValue = fmt.Sprintf(
				"%s\n%d. %s (<@%d>) - Level %d - Rank %d",
				fieldValue,
				rank,
				user.OSRSUsername,
				user.DiscordUserID,
				hiscore.Level,
				hiscore.Rank,
			)

			rank++
		}

		emojiName := types.NormalizeEmojiName(activity)

		activityEmoji, ok := types.ApplicationEmojis[emojiName]
		if !ok {
			activityEmoji = types.ApplicationEmojis["osrstrophy"]
		}

		emoji := fmt.Sprintf("<:%s>", activityEmoji.APIName())

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", emoji, activity),
			Value:  fieldValue,
			Inline: false,
		})

	}

	return fields, nil
}
