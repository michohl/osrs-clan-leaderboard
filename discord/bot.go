package discord

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/michohl/osrs-clan-leaderboard/storage"
	"github.com/michohl/osrs-clan-leaderboard/types"
	"github.com/michohl/osrs-clan-leaderboard/utils"
)

var (
	// BotToken is the token used for creating our bot
	BotToken = os.Getenv("DISCORD_BOT_TOKEN")
	// DevMode is a CSV list of servers to limit this code to. Used for local dev
	DevMode = os.Getenv("DEV_SERVERS")
)

func init() {
	types.BootstrapEmojis(BotToken)
}

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
		if DevMode != "" && !utils.IsServerInDevMode(server.ServerName) {
			log.Printf("dev mode is set and server '%s' is not present in dev whitelist. Not enabling cron schedule", server.ServerName)
			continue
		}
		EnableServerMessageCronjob(server, discord)
	}

	//discord.AddHandler(routeMessage)
	discord.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready) { log.Println("Bot is up!") })
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		// We can't use the state cache because it is
		// populated with all empty values
		guild, err := s.Guild(i.GuildID)
		if err != nil {
			log.Println(err)
			return
		}

		// If the development flag is set then only process
		// messages for servers that are whitelisted.
		if DevMode != "" && !utils.IsServerInDevMode(guild.Name) {
			log.Printf("dev mode is set and server '%s' is not present in dev whitelist. Skipping request", guild.Name)
			return
		}

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
			log.Printf("No handler for Interaction Type %s\n", i.Type)
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
