package reminder

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	interactionCustomIDParts = 4
	embedColorPink           = 0xFAC6DA
	hoursPerDay              = 24
	daysPerWeek              = 7
)

func (n *ReminderCmd) handleApplicationCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Printf("Processing /reminder command for user %s", i.Member.User.ID)
	m := generateModal(ReminderInfo{})
	if err := s.InteractionRespond(i.Interaction, m); err != nil {
		log.Printf("reminder: Failed to send modal for user %s: %v", i.Member.User.ID, err)
	}
}

func (n *ReminderCmd) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ModalSubmitData().CustomID != modalID {
		log.Printf("Invalid modal customID: %s for user %s", i.ModalSubmitData().CustomID, i.Member.User.ID)
		return
	}

	var rmdInfo ReminderInfo
	var rmdInfoExec ReminderInfoExec
	var validErr error

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

	rmdInfoExec.title = rmdInfo.title
	rmdInfoExec.eventTime, rmdInfoExec.triggerTime, validErr = rmdInfo.validate()
	if validErr != nil {
		rmdInfo.errMsg = validErr.Error()
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
	rmdInfoExec.executed = false

	customID := rmdInfo.generateCustomID()
	repository.PreHoldInfo(customID, rmdInfo)

	if validErr != nil {
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
		}
		return
	}

	repository.remindersInput.Delete(customID)
	if err := repository.HoldInfo(customID, rmdInfoExec); err != nil {
		log.Printf("reminder: Failed to persist reminder draft %s for user %s: %v", customID, i.Member.User.ID, err)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "エラー：リマインダーの保存に失敗しました。もう一度お試しください。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

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
	}
}

func (n *ReminderCmd) handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	if !strings.HasPrefix(customID, "reminder-") {
		log.Printf("Invalid button customID: %s for user %s", customID, i.Member.User.ID)
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラー：無効なボタン操作です。1",
		})
		if err != nil {
			log.Printf("reminder: Failed to send invalid customID response: %v", err)
		}
		return
	}

	parts := strings.Split(customID, "-")
	if len(parts) != interactionCustomIDParts {
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラー：無効なボタン操作です。2",
		})
		if err != nil {
			log.Printf("reminder: Failed to send malformed customID response: %v", err)
		}
		return
	}

	id := strings.Join(parts[:3], "-")
	action := parts[3]

	if action == "resend" {
		info, err := repository.PreLoad(id)
		if err != nil {
			log.Printf("No reminder data found for customID from reminderInfo: %s for user %s", id, i.Member.User.ID)
			_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "エラー：リマインダーデータが見つかりません。",
			})
			if err != nil {
				log.Printf("reminderInfo: Failed to send missing data response: %v", err)
			}
			return
		}
		repository.remindersInput.Delete(id)
		m := generateModal(info)
		if err := s.InteractionRespond(i.Interaction, m); err != nil {
			log.Printf("Failed to REsend modal for user %s: %v", i.Member.User.ID, err)
			return
		}
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("reminder: Failed to defer button response for user %s: %v", i.Member.User.ID, err)
		return
	}

	if action == "resendCancel" {
		log.Printf("Malformed button customID: %s for user %s", customID, i.Member.User.ID)
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "リマインダーの設定をキャンセルしました。",
		})
		if err != nil {
			log.Printf("reminder: Failed to send malformed customID response: %v", err)
		}
		return
	}

	infoExec, err := repository.Load(id)
	if err != nil {
		log.Printf("No reminder data found for customID from reminderInfoExec: %s for user %s", id, i.Member.User.ID)
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラー：リマインダーデータが見つかりません。",
		})
		if err != nil {
			log.Printf("reminderInfoExec: Failed to send missing data response: %v", err)
		}
		return
	}

	var response string
	if action == "cancel" {
		if err := repository.DeleteDraft(id); err != nil {
			log.Printf("reminder: Failed to cancel reminder %s for user %s: %v", id, i.Member.User.ID, err)
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "エラー：リマインダーのキャンセルに失敗しました。",
			})
			return
		}
		response = "リマインダーの設定をキャンセルしました。"
	} else if action == "confirm" {
		if err := repository.StoreInfo(id, infoExec); err != nil {
			log.Printf("reminder: Failed to confirm reminder %s for user %s: %v", id, i.Member.User.ID, err)
			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "エラー：リマインダーの確定に失敗しました。",
			})
			return
		}
		response = "リマインダーを確定しました。"
	} else {
		log.Printf("Unknown button action: %s for customID: %s for user %s", action, customID, i.Member.User.ID)
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラー：不明なボタン操作です。3",
		})
		if err != nil {
			log.Printf("reminder: Failed to send unknown action response: %v", err)
		}
		return
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: response,
	})
	if err != nil {
		log.Printf("reminder: Failed to send button response for user %s: %v", i.Member.User.ID, err)
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

func (n *ReminderCmd) errorMessageEmbed(errMsg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "入力を修正してください",
		Color: embedColorPink,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "エラー：", Value: errMsg, Inline: false},
		},
	}
}

func (n *ReminderCmd) confirmEmbed(rmdInfoExec ReminderInfoExec, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:     "リマインダーのための情報を取得しました",
		Color:     embedColorPink,
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
	weeks := totalHours / (hoursPerDay * daysPerWeek)
	days := (totalHours % (hoursPerDay * daysPerWeek)) / hoursPerDay
	hours := totalHours % hoursPerDay
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
