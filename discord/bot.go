package discord

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/storage"
)

var (
	// BotToken is the token used for creating our bot
	BotToken = os.Getenv("DISCORD_BOT_TOKEN")
	// ApplicationID is the ID for the application found in the Discord developer portal
	ApplicationID = os.Getenv("DISCORD_APP_ID")
)

// StartBotListener is the function that starts the Discord
// bot and makes it available to accept commands from users
func StartBotListener() {
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal(err)
	}

	allServers, err := storage.FetchAllServers()
	if err != nil {
		panic(err)
	}

	// Re-enable any crons configured in our database
	for _, server := range allServers {
		EnableServerMessageCronjob(server, discord)
	}

	BootstrapEmojis(discord)

	//discord.AddHandler(routeMessage)
	discord.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready) { log.Println("Bot is up!") })
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			commandFunction := GetCommandHandler(i.ApplicationCommandData().Name)
			if commandFunction != nil {
				commandFunction(s, i)
			}
		case discordgo.InteractionModalSubmit:
			modalSubmitFunction := GetModalSubmitHandler(i.ModalSubmitData().CustomID)
			if modalSubmitFunction != nil {
				modalSubmitFunction(s, i)
			}
		default:
			log.Fatalf("No handler for Interaction Type %s", i.Type)
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
