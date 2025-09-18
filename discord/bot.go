package discord

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// StartBotListener is the function that starts the Discord
// bot and makes it available to accept commands from users
func StartBotListener() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	discord, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}

	//discord.AddHandler(routeMessage)
	discord.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready) { log.Println("Bot is up!") })
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// open session
	discord.Open()
	defer discord.Close() // close session, after function termination

	// Register commands for auto completion
	var GuildID string
	_, err = discord.ApplicationCommandBulkOverwrite(
		discord.State.User.ID,
		GuildID,
		commands,
	)
	if err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	// keep bot running untill there is NO os interruption (ctrl + C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
