package reminder

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
	if len(parts) != 4 {
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラー：無効なボタン操作です。2",
		})
		if err != nil {
			log.Printf("reminder: Failed to send malformed customID response: %v", err)
		}
		return
	}

	id := strings.Join(parts[:3], "-") // reminder-YYYYMMDDHHMM-RAND
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

	// Immediately defer response to prevent timeout
	//must not come before modal resending
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
		repository.reminders.Delete(id)
		response = "リマインダーの設定をキャンセルしました。"
	} else if action == "confirm" {
		// Simulate DB save (log for now as DB is not ready)
		repository.StoreInfo(id, infoExec)
		repository.reminders.Delete(id)
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
