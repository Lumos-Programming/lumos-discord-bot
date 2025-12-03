package reminder

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
	var rmdInfoExec ReminderInfoExec
	var validErr error

	//input from user
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
	rmdInfo.errCode = []int{0, 0, 0}
	rmdInfo.errMsg = ""

	// Validate and caliculate input
	rmdInfoExec.title = rmdInfo.title
	rmdInfoExec.eventTime, rmdInfoExec.triggerTime, validErr = rmdInfo.validate()
	log.Printf("returned to modal_submit: finish validate with errCode%v", rmdInfo.errCode)
	if validErr != nil {
		rmdInfo.errMsg = validErr.Error()
	} else {
		rmdInfo.errMsg = ""
	}
	rmdInfoExec.UserName = i.Member.Nick
	if rmdInfoExec.UserName == "" {
		rmdInfoExec.UserName = i.Member.User.GlobalName
	}
	if rmdInfoExec.UserName == "" {
		rmdInfoExec.UserName = i.Member.User.Username
	}
	rmdInfoExec.UserID = i.Member.User.ID
	rmdInfoExec.ChannelID = i.ChannelID
	rmdInfoExec.Session = s
	rmdInfoExec.executed = false

	// Generate custom ID
	customID := rmdInfo.generateCustomID()
	repository.PreHoldInfo(customID, rmdInfo)
	log.Printf("preHolded rmdInfo with customID: %s for user %s", customID, i.Member.User.ID)

	//Send error message and request correction
	if validErr != nil {
		log.Printf("Validation failed for user %s: %s", i.Member.User.ID, rmdInfo.errMsg)
		embed := n.errorMessageEmbed(rmdInfo.errMsg)
		resendButton := discordgo.Button{
			Label:    "再入力",
			Style:    discordgo.PrimaryButton,
			CustomID: customID + "-resend",
		}
		cancelButton := discordgo.Button{
			Label:    "キャンセル",
			Style:    discordgo.SecondaryButton,
			CustomID: customID + "-resendCancel",
		}
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{cancelButton, resendButton},
					},
				},
			},
		}); err != nil {
			log.Printf("reminder: Failed to send error message for user %s: %v", i.Member.User.ID, err)
		} else {
			log.Printf("Sent error message with buttons for user %s", i.Member.User.ID)
		}
		return
	}

	repository.remindersInput.Delete(customID)
	log.Printf("Deleted preHolded rmdInfo")
	repository.HoldInfo(customID, rmdInfoExec)
	log.Printf("Stored reminder with customID: %s for user %s", customID, i.Member.User.ID)

	// Send confirmation message with buttons
	embed := n.confirmEmbed(rmdInfoExec, i)
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

func (n *ReminderCmd) errorMessageEmbed(errMsg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "入力を修正してください",
		Color: 0xFAC6DA, // pink
		//Timestamp: time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "エラー：", Value: errMsg, Inline: false},
		},
	}
}

func (n *ReminderCmd) confirmEmbed(rmdInfoExec ReminderInfoExec, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:     "リマインダーのための情報を取得しました",
		Color:     0xFAC6DA, // pink
		Timestamp: time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "イベント名", Value: rmdInfoExec.title, Inline: false},
			{Name: "開催日時", Value: fmt.Sprintf("%s", rmdInfoExec.eventTime), Inline: false},
			{Name: "リマインダーのタイミング", Value: invertSetTime(rmdInfoExec), Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("%s", rmdInfoExec.UserName),
			IconURL: i.Member.User.AvatarURL("64"),
		},
	}
}

func invertSetTime(rmdInfoExec ReminderInfoExec) string {
	stTime := rmdInfoExec.eventTime.Sub(rmdInfoExec.triggerTime)

	totalHours := int(stTime.Hours())
	weeks := totalHours / (24 * 7)
	days := (totalHours % (24 * 7)) / 24
	hours := totalHours % 24
	minutes := int(stTime.Minutes()) % 60

	var parts []string
	if weeks > 0 {
		parts = append(parts, fmt.Sprintf("%d週間", weeks))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d日", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d時間", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d分", minutes))
	}

	if len(parts) == 0 {
		parts = append(parts, "0分")
	}

	result := strings.Join(parts, "")
	return result + "前"
}
