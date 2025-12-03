package reminder

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (n *ReminderCmd) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Printf("Processing /reminder command for user %s", i.Member.User.ID)
	m := generateModal(ReminderInfo{})
	if err := s.InteractionRespond(i.Interaction, m); err != nil {
		log.Printf("reminder: Failed to send modal for user %s: %v", i.Member.User.ID, err)
	}
}

func generateModal(prevInfo ReminderInfo) *discordgo.InteractionResponse {
	modalTitle := "大切なイベントのリマインダーを設定しましょう!"
	if prevInfo.errMsg != "" {
		if prevInfo.errCode[0] == 1 {
			prevInfo.eventYear = ""
		}
		if prevInfo.errCode[1] == 1 {
			prevInfo.eventTime = ""
		}
		if prevInfo.errCode[2] == 1 {
			prevInfo.setTime = ""
		}
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    titleID,
					Label:       "イベントの名前",
					Style:       discordgo.TextInputShort,
					Placeholder: "e.g. 'Security Trend Share Project'",
					Required:    true,
					Value:       prevInfo.title,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    yearID,
					Label:       "開催年(今年ではない場合)",
					Style:       discordgo.TextInputShort,
					Placeholder: "e.g. '2026'",
					Required:    false,
					Value:       prevInfo.eventYear,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    timeID,
					Label:       "開催日時",
					Style:       discordgo.TextInputShort,
					Placeholder: "e.g. '12012300' (12/1 23:00)",
					Required:    true,
					Value:       prevInfo.eventTime,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    setID,
					Label:       "リマインダーのタイミング",
					Style:       discordgo.TextInputShort,
					Placeholder: "e.g. '1h30m' (week: 'w', day: 'd', hour: 'h', minute: 'm')",
					Required:    true,
					Value:       prevInfo.setTime,
				},
			},
		},
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   modalID,
			Title:      modalTitle,
			Components: components,
		},
	}
}
