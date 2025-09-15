package main

import (
	"fmt"
	"log"
	"lumos-discord-bot/cmd"
	del "lumos-discord-bot/cmd/delete"
	"lumos-discord-bot/cmd/nox"
	"lumos-discord-bot/cmd/reminder"
	tryhackme_achievement "lumos-discord-bot/cmd/tryhackme-achievement"
	"lumos-discord-bot/handler"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	EnvDiscordToken       = "DISCORD_TOKEN"
	EnvTargetServer       = "TARGET_SERVER"
	EnvWelcomeChannel     = "WELCOME_CHANNEL"
	EnvMemberRoleID       = "MEMBER_ROLE_ID"
	EnvTestUserID         = "TEST_USER_ID"
	EnvTHMAchievementChan = "TRYHACKME_ACHIEVEMENT_CHANNEL_ID"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env file not found")
	}
	discordToken := os.Getenv(EnvDiscordToken)
	welcomeChannel := os.Getenv(EnvWelcomeChannel)
	targetServer := os.Getenv(EnvTargetServer)
	memberRoleID := os.Getenv(EnvMemberRoleID)
	testUserID := os.Getenv(EnvTestUserID)
	thmAchievementChannelID := os.Getenv(EnvTHMAchievementChan)

	bot, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		panic(err)
	}

	bot.Identify.Intents = discordgo.IntentsAll

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	if err = bot.Open(); err != nil {
		panic(err)
	}

	// setup commands...
	cmds := cmd.NewExec()

	noxCmd := nox.NewNoxCmd()
	cmds.Add(noxCmd)
	deleteCmd := del.NewDeleteCmd()
	cmds.Add(deleteCmd)
	tryhackmeAchievementCmd := tryhackme_achievement.NewTryHackMeCmd(targetServer, thmAchievementChannelID)
	cmds.Add(tryhackmeAchievementCmd)
	reminderCmd := reminder.NewReminderCmd()
	cmds.Add(reminderCmd)

	cmdHandler := cmds.Activate(bot)
	defer cmdHandler.Deactivate()
	// setup commands end

	welcome := handler.NewWelcomeHandler(targetServer, welcomeChannel, memberRoleID, testUserID)
	bot.AddHandler(welcome.Handle)

	defer bot.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
