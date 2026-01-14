package reminder

import (
	"sync/atomic"

	"github.com/bwmarrin/discordgo"
)

type DiscordSender interface {
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

type senderHolder struct {
	s DiscordSender
}

var sender atomic.Value // senderHolder

func init() {
	sender.Store(senderHolder{})
}

func SetDiscordSender(s DiscordSender) {
	sender.Store(senderHolder{s: s})
}

func getDiscordSender() DiscordSender {
	return sender.Load().(senderHolder).s
}
