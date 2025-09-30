package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// GetChannel is a helper function that returns a discord channel the "hard way" meaning we don't
// attempt to use the cache and instead query discord for all the channels and return the one that matches
// what the user asked for
func GetChannel(s *discordgo.Session, guildID string, channelName string) (*discordgo.Channel, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	var channel *discordgo.Channel = &discordgo.Channel{}
	for _, c := range channels {
		if c.Name == channelName {
			channel = c
			break
		}

	}

	if channel.Name == "" {
		return nil, fmt.Errorf("Unable to find channel %s in guild %s", channelName, guildID)
	}

	return channel, nil
}
