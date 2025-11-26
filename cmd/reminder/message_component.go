package reminder

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (n *ReminderCmd) handleMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	log.Printf("Processing button interaction with customID: %s for user %s", customID, i.Member.User.ID)

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
		log.Printf("Malformed button customID: %s for user %s", customID, i.Member.User.ID)
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
		m := generateModal("", ReminderInfo{})
		if err := s.InteractionRespond(i.Interaction, m); err != nil {
			log.Printf("Failed to REsend modal for user %s: %v", i.Member.User.ID, err)
			return
		} else {
			log.Printf("REsent modal for user %s", i.Member.User.ID)
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

	info, err := repository.Load(id)
	if err != nil {
		log.Printf("No reminder data found for customID: %s for user %s", id, i.Member.User.ID)
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "エラー：リマインダーデータが見つかりません。",
		})
		if err != nil {
			log.Printf("reminder: Failed to send missing data response: %v", err)
		}
		return
	}

	var response string
	if action == "cancel" {
		repository.reminders.Delete(id)
		response = "リマインダーの設定をキャンセルしました。"
		log.Printf("Cancelled reminder with customID: %s for user %s", id, i.Member.User.ID)
	} else if action == "confirm" {
		// Simulate DB save (log for now as DB is not ready)
		log.Printf("Saving to DB for user %s: {title:%s eventTime:%s triggerTime:%s executed:%t}", i.Member.User.ID, info.title, info.eventTime.String(), info.triggerTime.String(), info.executed)
		repository.StoreInfo(id, info)
		repository.reminders.Delete(id)
		response = "リマインダーを確定しました。"
		log.Printf("Confirmed reminder with customID: %s for user %s", id, i.Member.User.ID)
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
