package nox

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
)

var greetingMessages = []string{
	"お疲れ!", "Goodbye! See you again!!", "See ya!", "またな!", "printf(\"Goodbye, world!\");", "続けるにはENTERを押すかコマンドを入力してください", "terminated with status code 0", "Console.WriteLine(\"Goodbye, world!\");",
}

type NoxCmd struct {
}

func NewNoxCmd() NoxCmd {
	return NoxCmd{}
}

func (n NoxCmd) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "nox",
		Description: "疲れたのかい？ならばnoxだ！",
	}
}

func (n NoxCmd) Handle(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	// get member's voice state

	// select random message
	r := rand.Intn(len(greetingMessages))
	greetMes := greetingMessages[r]

	name := i.Member.Nick
	if name == "" {
		name = i.Member.User.Username
	}
	mes := discordgo.MessageEmbed{
		Color:       0xF1C40F,
		Footer:      &discordgo.MessageEmbedFooter{Text: greetMes},
		Description: fmt.Sprintf("%s さんが退出します!  <@%s>", i.Member.User.Username, i.Member.User.ID),
		Author: &discordgo.MessageEmbedAuthor{
			Name:    name,
			URL:     fmt.Sprintf("https://discordapp.com/users/%s", i.Member.User.ID),
			IconURL: i.Member.User.AvatarURL(""),
		},
	}

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
}
