package types

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// ApplicationEmojis is a collection of all of the app's personal
// emojis in a map where the key is the human
// readable name for the emoji. e.g. "woodcutting"
var (
	ApplicationEmojis = map[string]*discordgo.Emoji{}

	// ApplicationID is the ID for the application found in the Discord developer portal
	ApplicationID = os.Getenv("DISCORD_APP_ID")
)

// BootstrapEmojis is responsible for populating our static list
// of all the application's available emojis
func BootstrapEmojis(botToken string) error {
	discord, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}

	emojis, err := discord.ApplicationEmojis(ApplicationID)
	if err != nil {
		return err
	}

	for _, emoji := range emojis {
		ApplicationEmojis[emoji.Name] = emoji
	}

	return nil
}

// NormalizeEmojiName takes an acitivity or skill
// name as it would be discovered from the API and converts
// it into a valid discord emoji name
// e.g. Kree'Arra -> kreearra
//
//	Commander Zilyana -> commander_zilyana
//	Clue Scrolls (beginner) -> clue_scrolls
func NormalizeEmojiName(name string) string {

	name = strings.Trim(name, " ")

	weirdCases := map[string]string{
		"Bounty Hunter ":    "skulled",
		"Clue Scrolls ":     "clue_scroll",
		"Chambers of Xeric": "cox",
		"Theatre of Blood":  "tob",
		"Tombs of Amascut":  "toa",
		"Nightmare":         "nightmare",
		"PvP":               "skulled",
		"LMS":               "skulled",
		"Rifts closed":      "gotr",
	}

	// Some activities have "variations" that we want to all resolve to the same thing
	for p, n := range weirdCases {
		if strings.Contains(name, p) {
			return n
		}
	}

	for _, c := range []string{"'", "-", ":"} {
		name = strings.ReplaceAll(name, c, "")
	}

	name = strings.ReplaceAll(name, " ", "_")

	name = strings.ToLower(name)

	return name
}
