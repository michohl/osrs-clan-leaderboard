package discord

import "github.com/bwmarrin/discordgo"

// ApplicationEmojis is a collection of all of the app's personal
// emojis in a map where the key is the human
// readable name for the emoji. e.g. "woodcutting"
var (
	ApplicationEmojis = map[string]*discordgo.Emoji{}
)

// BootstrapEmojis is responsible for populating our static list
// of all the application's available emojis
func BootstrapEmojis(s *discordgo.Session) error {
	emojis, err := s.ApplicationEmojis(ApplicationID)
	if err != nil {
		return err
	}

	for _, emoji := range emojis {
		ApplicationEmojis[emoji.Name] = emoji
	}

	return nil
}
