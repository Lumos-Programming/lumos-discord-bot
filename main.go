package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"os/signal"
)

const (
	ENV_DISCORD_TOKEN   = "DISCORD_TOKEN"
	ENV_WELCOME_CHANNEL = "WELCOME_CHANNEL"
)

type Command struct {
	discordgo.ApplicationCommand
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var greetingMessages = []string{
	"お疲れ!", "Goodbye! See you again!!", "See ya!", "またな!", "printf(\"Goodbye, world!\");", "続けるにはENTERを押すかコマンドを入力してください", "terminated with status code 0", "Console.WriteLine(\"Goodbye, world!\");",
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	discordToken := os.Getenv(ENV_DISCORD_TOKEN)
	//welcomeChannel := os.Getenv(ENV_WELCOME_CHANNEL)

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

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "nox",
			Description: "疲れたのかい？ならばnoxだ！",
		},
	}

	// first, delete all existing commands
	existingCommands, err := bot.ApplicationCommands(bot.State.User.ID, "")
	if err != nil {
		log.Printf("failed to get commands, err: %v", err)
	}
	for _, v := range existingCommands {
		err := bot.ApplicationCommandDelete(bot.State.User.ID, "", v.ID)
		if err != nil {
			log.Printf("failed to delete command: %v, err: %v", v.Name, err)
		} else {
			log.Printf("deleted command: %v", v.Name)
		}
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, v := range commands {
		// if guildID is empty, it will create a global command
		newCmd, err := bot.ApplicationCommandCreate(bot.State.User.ID, "", v)
		if err != nil {
			log.Printf("failed to create command: %v, err: %v", v.Name, err)
		} else {
			log.Printf("created command: %v", newCmd.Name)
		}
		registeredCommands = append(registeredCommands, newCmd)
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"nox": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// get member's voice state
			// select random message
			r := rand.Intn(len(greetingMessages))
			greetMes := greetingMessages[r]

			// create message to welcome channel
			// first create embed message
			mes := discordgo.MessageEmbed{
				Color:       0xF1C40F,
				Footer:      &discordgo.MessageEmbedFooter{Text: greetMes},
				Description: fmt.Sprintf("%s さんが退出します!  <@%s>", i.Member.User.Username, i.Member.User.ID),
				Author: &discordgo.MessageEmbedAuthor{
					Name:    i.Member.Nick,
					URL:     i.Member.User.AvatarURL(""),
					IconURL: i.Member.User.AvatarURL(""),
				},
			}

			// send embed message to welcome channel
			_, err := s.ChannelMessageSendEmbed(i.ChannelID, &mes)
			if err != nil {
				log.Printf("failed to send message, err: %v", err)
			}

			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Noxコマンドを発動しました! お疲れ!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); err != nil {
				log.Printf("failed to send followup message, err: %v", err)
			}
		},
	}

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			if h != nil {
				h(s, i)
			} else {
				log.Printf("no handler for command: %v", i.ApplicationCommandData().Name)
			}
		}
	})

	defer bot.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
