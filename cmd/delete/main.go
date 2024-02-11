package delete

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type DeleteCmd struct {
}

func NewDeleteCmd() DeleteCmd {
	return DeleteCmd{}
}

func (n DeleteCmd) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "delete",
		Description: "先頭から指定した数のメッセージを削除します。",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "count",
				Description: "削除するメッセージの数",
				Required:    true,
			},
		},
	}
}

func (n DeleteCmd) Handle(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	opts := i.ApplicationCommandData().Options
	if len(opts) != 1 {
		log.Printf("invalid options: %#v", opts)
		return
	}
	count := opts[0].IntValue()
	if count < 1 {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "件数が無効です。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			log.Printf("failed to respond, err: %v", err)
		}
		return
	} else if count > 30 {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "削除件数が多すぎます。30件以下にしてください。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}); err != nil {
			log.Printf("failed to respond, err: %v", err)
		}
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		log.Printf("failed to do interaction response, err: %v", err)
		return
	}

	// create follow-up message
	followUp := discordgo.WebhookParams{
		Content: "削除中...",
		Flags:   discordgo.MessageFlagsEphemeral,
	}
	followUpMsg, err := s.FollowupMessageCreate(i.Interaction, true, &followUp)
	if err != nil {
		log.Printf("failed to send follow-up message, err: %v", err)
		return
	}

	// delete messages
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		return
	}
	messages, err := s.ChannelMessages(
		channel.ID, int(count), "", "", "")
	if err != nil {
		log.Printf("failed to get messages, err: %v", err)
		return
	}

	mesIDs := make([]string, 0, len(messages))
	for i := range messages {
		mesIDs = append(mesIDs, messages[i].ID)
	}

	if err := s.ChannelMessagesBulkDelete(channel.ID, mesIDs); err != nil {
		log.Printf("failed to delete messages, err: %v", err)
	}

	finishFollowUpStr := fmt.Sprintf("%d件のメッセージを削除しました。", len(mesIDs))
	finishFollowUp := discordgo.WebhookEdit{
		Content: &finishFollowUpStr,
	}
	if _, err := s.FollowupMessageEdit(i.Interaction, followUpMsg.ID, &finishFollowUp); err != nil {
		log.Printf("failed to edit follow-up message, err: %v", err)
		return
	}
}
