package reminder

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (n *ReminderCmd) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ModalSubmitData().CustomID != modalID {
		log.Printf("Invalid modal customID: %s for user %s", i.ModalSubmitData().CustomID, i.Member.User.ID)
		return
	}
	log.Printf("Processing modal submit for user %s", i.Member.User.ID)

	var rmdInfo ReminderInfo
	for _, cmp := range i.ModalSubmitData().Components {
		row := cmp.(*discordgo.ActionsRow)
		for _, inner := range row.Components {
			input := inner.(*discordgo.TextInput)
			switch input.CustomID {
			case titleID:
				rmdInfo.title = input.Value
			case yearID:
				rmdInfo.eventYear = input.Value
				if rmdInfo.eventYear == "" {
					rmdInfo.eventYear = strconv.Itoa(time.Now().Year())
				}
			case timeID:
				rmdInfo.eventTime = input.Value
			case setID:
				rmdInfo.setTime = input.Value
			}
		}
	}
	rmdInfo.UserID = i.Member.User.ID
	rmdInfo.ChannelID = i.ChannelID
	rmdInfo.Session = s

	// Validate input
	err := rmdInfo.validate()
	if err != nil {
		errorMsg := "正しい形式で入力してください: " + err.Error() // 未修正：ログではなく、モーダルに表示させたい
		log.Printf("Validation failed for user %s: %s", i.Member.User.ID, errorMsg)
		generateModal(errorMsg, rmdInfo)
		return
	}

	// Generate custom ID
	customID := rmdInfo.generateCustomID()
	repository.HoldInfo(customID, rmdInfo)
	log.Printf("Stored reminder with customID: %s for user %s", customID, i.Member.User.ID)

	// Send confirmation message with buttons
	embed := n.confirmEmbed(rmdInfo, i)
	cancelButton := discordgo.Button{
		Label:    "キャンセル",
		Style:    discordgo.SecondaryButton,
		CustomID: customID + "-cancel",
	}
	confirmButton := discordgo.Button{
		Label:    "確定",
		Style:    discordgo.PrimaryButton,
		CustomID: customID + "-confirm",
	}
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{cancelButton, confirmButton},
				},
			},
		},
	}); err != nil {
		log.Printf("reminder: Failed to send confirmation message for user %s: %v", i.Member.User.ID, err)
	} else {
		log.Printf("Sent confirmation message with buttons for user %s", i.Member.User.ID)
	}

}

func (n *ReminderCmd) confirmEmbed(rmdInfo ReminderInfo, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	name := i.Member.Nick
	if name == "" {
		name = i.Member.User.GlobalName
	}
	if name == "" {
		name = i.Member.User.Username
	}
	eventYear := rmdInfo.eventYear
	eventMonth := rmdInfo.eventTime[:2]
	eventDay := rmdInfo.eventTime[2:4]
	eventHour := rmdInfo.eventTime[4:6]
	eventMinute := rmdInfo.eventTime[6:8]
	return &discordgo.MessageEmbed{
		Title:     "リマインダーのための情報を取得しました",
		Color:     0xFAC6DA, // pink
		Timestamp: time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "イベント名", Value: rmdInfo.title, Inline: false},
			{Name: "開催日時", Value: eventYear + "/" + eventMonth + "/" + eventDay + " " + eventHour + ":" + eventMinute, Inline: false},
			{Name: "リマインダーのタイミング", Value: invertSetTime(rmdInfo), Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("%s (@%s)", name, i.Member.User.Username),
			IconURL: i.Member.User.AvatarURL("64"),
		},
	}
}

func invertSetTime(rmdInfo ReminderInfo) string {
	message := ""
	s := rmdInfo.setTime
	i := 0
	for i < len(s) {
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			message = message + s[i:i+1]
			i++
		}
		unit := s[i]
		i++
		switch unit {
		case 'w':
			message += "週間"
		case 'd':
			message += "日"
		case 'h':
			message += "時間"
		case 'm':
			message += "分"
		default:
			return ""
		}
	}
	return message + "前"
}
