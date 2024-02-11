package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"lumos-discord-bot/cmd"
	"lumos-discord-bot/cmd/nox"
	"lumos-discord-bot/handler"
	"os"
	"os/signal"
)

const (
	EnvDiscordToken   = "DISCORD_TOKEN"
	EnvWelcomeChannel = "WELCOME_CHANNEL"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env is not loaded")
	}
	discordToken := os.Getenv(EnvDiscordToken)
	welcomeChannel := os.Getenv(EnvWelcomeChannel)

	bot, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		panic(err)
	}
	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	if err = bot.Open(); err != nil {
		panic(err)
	}

	cmds := cmd.NewExec()

	noxCmd := nox.NewNoxCmd()
	cmds.Add(noxCmd)

	bot.AddHandler(cmds.Handle)

	welcome := handler.NewWelcomeHandler(welcomeChannel)
	bot.AddHandler(welcome.Handle)

	defer bot.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
