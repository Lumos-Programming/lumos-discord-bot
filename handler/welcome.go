package handler

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type WelcomeHandler struct {
	channel string
}

func NewWelcomeHandler(channel string) WelcomeHandler {
	return WelcomeHandler{channel: channel}
}

func (h *WelcomeHandler) Handle(s *discordgo.Session, i *discordgo.GuildMemberAdd) {
	embedMes := discordgo.MessageEmbed{
		Title:       "Lumosへようこそ!!!",
		Description: fmt.Sprintf("<@%s> さんがLumosにやってきました:sparkles:", i.User.ID),
		Color:       0xF1C40F,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    i.User.Username,
			URL:     fmt.Sprintf("https://discordapp.com/users/%s", i.User.ID),
			IconURL: i.User.AvatarURL(""),
		},
	}
	_, err := s.ChannelMessageSendEmbed(h.channel, &embedMes)
	if err != nil {
		log.Printf("failed to send message, err: %v", err)
	}
}
