package reminder

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"time"
)

const (
	modalID = "reminder"
	titleID = "plan-title"
	yearID  = "plan-year"
	timeID  = "plan-time"
	setID   = "reminder-set-time"
)

type ReminderCmd struct{}

func NewReminderCmd() *ReminderCmd {
	return &ReminderCmd{}
}

func (n *ReminderCmd) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		modal := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: modalID,
				Title:    "大切なイベントのリマインダーを設定しましょう!",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    titleID,
								Label:       "イベントの名前",
								Style:       discordgo.TextInputShort,
								Placeholder: "e.g. 'Security Trend Share Project'",
								Required:    true,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    yearID,
								Label:       "開催年(デフォルトでは今年)",
								Style:       discordgo.TextInputShort,
								Placeholder: "e.g. '2026'",
								Required:    false,
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
							},
						},
					},
				},
			},
		}

		if err := s.InteractionRespond(i.Interaction, modal); err != nil {
			log.Printf("reminder: embed respond error: %v", err)
		}

	case discordgo.InteractionModalSubmit:
		if i.ModalSubmitData().CustomID != modalID {
			return
		}
		var title, eventYear, eventTime, setTime string
		for _, cmp := range i.ModalSubmitData().Components {
			row := cmp.(*discordgo.ActionsRow)
			for _, inner := range row.Components {
				input := inner.(*discordgo.TextInput)
				switch input.CustomID {
				case titleID:
					title = input.Value
				case yearID:
					eventYear = input.Value
					if eventYear == "" {
						eventYear = strconv.Itoa(time.Now().Year())
					}
				case timeID:
					eventTime = input.Value
				case setID:
					setTime = input.Value
				}
			}
		}

		embed := &discordgo.MessageEmbed{
			Title:     "リマインダーのための情報を取得しました",
			Color:     0xFAC6DA, //pink
			Timestamp: time.Now().Format(time.RFC3339),
			Fields: []*discordgo.MessageEmbedField{
				{Name: "イベント名", Value: title, Inline: false},
				{Name: "開催年", Value: eventYear, Inline: true},
				{Name: "開催日時", Value: eventTime, Inline: true},
				{Name: "リマインダーのタイミング", Value: setTime, Inline: false},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    fmt.Sprintf("%s<@%s>", i.Member.Nick, i.Member.User.Username),
				IconURL: i.Member.User.AvatarURL("64"),
			},
		}
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		}); err != nil {
			log.Printf("reminder: embed respond error: %v", err)
		}
	}
}

func (n *ReminderCmd) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "reminder",
		Description: "大切な予定をn分前にお知らせします！",
	}
}

func (n *ReminderCmd) ModalCustomIDs() []string {
	return []string{modalID}
}
