package reminder

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	modalID = "reminder"
	titleID = "plan-title"
	yearID  = "plan-year"
	timeID  = "plan-time"
	setID   = "reminder-set-time"
	aMinute = time.Minute
	anHour  = time.Hour
	aDay    = time.Hour * 24
	aWeek   = aDay * 7
	aYear   = aDay * 365
)

type ReminderInfo struct {
	title     string
	eventYear string
	eventTime string
	setTime   string
}

var reminders sync.Map // map[string]ReminderInfo for temporary storage

type ReminderCmd struct{}

func NewReminderCmd() *ReminderCmd {
	return &ReminderCmd{}
}

func (n *ReminderCmd) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Printf("Entering ReminderCmd.Handle: Type=%s, InteractionID=%s, UserID=%s", i.Type, i.ID, i.Member.User.ID)
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		log.Printf("Processing /reminder command for user %s", i.Member.User.ID)
		n.showModal(s, i, "", ReminderInfo{})
	case discordgo.InteractionModalSubmit:
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

		// Validate input
		valid, errs := validateReminder(rmdInfo)
		if !valid {
			errorMsg := "正しい形式で入力してください: " + strings.Join(errs, ", ") // 未修正：ログではなく、モーダルに表示させたい
			log.Printf("Validation failed for user %s: %s", i.Member.User.ID, errorMsg)
			n.showModal(s, i, errorMsg, rmdInfo)
			return
		}

		// Generate custom ID
		now := time.Now()                      //未修正：リマインダーを設定した日の日付ではなく、リマインダー対象のイベントの日付にしたい
		dateStr := now.Format("20060102-1504") // YYYYMMDD-HHMM
		randNum := rand.Intn(10000)
		randStr := fmt.Sprintf("%04d", randNum)
		customID := fmt.Sprintf("reminder-%s-%s", dateStr, randStr)
		reminders.Store(customID, rmdInfo)
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
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		log.Printf("Processing button interaction with customID: %s for user %s", customID, i.Member.User.ID)

		// Immediately defer response to prevent timeout
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		}); err != nil {
			log.Printf("reminder: Failed to defer button response for user %s: %v", i.Member.User.ID, err)
			return
		}

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
		if len(parts) != 5 {
			log.Printf("Malformed button customID: %s for user %s", customID, i.Member.User.ID)
			_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "エラー：無効なボタン操作です。2",
			})
			if err != nil {
				log.Printf("reminder: Failed to send malformed customID response: %v", err)
			}
			return
		}

		id := strings.Join(parts[:4], "-") // reminder-YYYYMMDD-HHMM-RAND
		action := parts[4]
		infoI, ok := reminders.Load(id)
		if !ok {
			log.Printf("No reminder data found for customID: %s for user %s", id, i.Member.User.ID)
			_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "エラー：リマインダーデータが見つかりません。",
			})
			if err != nil {
				log.Printf("reminder: Failed to send missing data response: %v", err)
			}
			return
		}

		info := infoI.(ReminderInfo)
		var response string
		if action == "cancel" {
			reminders.Delete(id)
			response = "リマインダーの設定をキャンセルしました。"
			log.Printf("Cancelled reminder with customID: %s for user %s", id, i.Member.User.ID)
		} else if action == "confirm" {
			// Simulate DB save (log for now as DB is not ready)
			log.Printf("Saving to DB for user %s: %+v", i.Member.User.ID, info)
			reminders.Delete(id)
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

		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: response,
		})
		if err != nil {
			log.Printf("reminder: Failed to send button response for user %s: %v", i.Member.User.ID, err)
		}
	}
	log.Printf("Exiting ReminderCmd.Handle for user %s", i.Member.User.ID)
}

func (n *ReminderCmd) showModal(s *discordgo.Session, i *discordgo.InteractionCreate, errorMsg string, prevInfo ReminderInfo) {
	modalTitle := "大切なイベントのリマインダーを設定しましょう!"
	if errorMsg != "" {
		modalTitle = errorMsg
		if len(modalTitle) > 45 { // Discord modal title limit
			modalTitle = modalTitle[:45]
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

	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID:   modalID,
			Title:      modalTitle,
			Components: components,
		},
	}

	if err := s.InteractionRespond(i.Interaction, modal); err != nil {
		log.Printf("reminder: Failed to send modal for user %s: %v", i.Member.User.ID, err)
	} else {
		log.Printf("Sent modal for user %s", i.Member.User.ID)
	}
}

func validateReminder(info ReminderInfo) (bool, []string) {
	errs := []string{}

	// Validate title (non-empty)
	if info.title == "" {
		errs = append(errs, "イベント名は必須です") // 必須入力の確認はDiscordがやってくれるので必要ないと思う
	}

	// Validate eventYear
	if info.eventYear != "" {
		y, err := strconv.Atoi(info.eventYear)
		if err != nil || len(info.eventYear) != 4 || y < time.Now().Year() {
			errs = append(errs, "開催年は4桁の数字で、今年以降にしてください")
		}
	}

	// Validate eventTime
	if len(info.eventTime) != 8 || !isAllDigits(info.eventTime) {
		errs = append(errs, "開催日時は8桁の数字 (MMDDHHmm) にしてください")
	} else {
		month, _ := strconv.Atoi(info.eventTime[:2])
		day, _ := strconv.Atoi(info.eventTime[2:4])
		hour, _ := strconv.Atoi(info.eventTime[4:6])
		min, _ := strconv.Atoi(info.eventTime[6:8])
		if month < 1 || month > 12 || day < 1 || day > 31 || hour < 0 || hour > 23 || min < 0 || min > 59 {
			errs = append(errs, "開催日時の値が無効です")
		}
	}

	// Validate setTime
	if _, err := parseCustomDuration(info.setTime); err != nil {
		errs = append(errs, "リマインダーのタイミングは '1w2d3h4m' 形式にしてください")
	}

	return len(errs) == 0, errs
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parseCustomDuration(s string) (time.Duration, error) {
	var d time.Duration
	i := 0
	for i < len(s) {
		num := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			num = num*10 + int(s[i]-'0')
			i++
		}
		if i == len(s) {
			return 0, fmt.Errorf("invalid format")
		}
		unit := s[i]
		i++
		switch unit {
		case 'w':
			d += time.Duration(num) * aWeek
		case 'd':
			d += time.Duration(num) * aDay
		case 'h':
			d += time.Duration(num) * anHour
		case 'm':
			d += time.Duration(num) * aMinute
		default:
			return 0, fmt.Errorf("invalid unit: %c", unit)
		}
	}
	return d, nil
}

func (n *ReminderCmd) confirmEmbed(rmdInfo ReminderInfo, i *discordgo.InteractionCreate) *discordgo.MessageEmbed {
	name := i.Member.Nick
	if name == "" {
		name = i.Member.User.GlobalName
	}
	if name == "" {
		name = i.Member.User.Username
	}
	return &discordgo.MessageEmbed{
		Title:     "リマインダーのための情報を取得しました",
		Color:     0xFAC6DA, // pink
		Timestamp: time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "イベント名", Value: rmdInfo.title, Inline: false},
			{Name: "開催年", Value: rmdInfo.eventYear, Inline: true},
			{Name: "開催日時", Value: rmdInfo.eventTime, Inline: true},
			{Name: "リマインダーのタイミング", Value: rmdInfo.setTime, Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("%s (@%s)", name, i.Member.User.Username),
			IconURL: i.Member.User.AvatarURL("64"),
		},
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
