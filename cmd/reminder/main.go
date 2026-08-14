package reminder

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
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

var repository ReminderRepository

type ReminderRepository struct {
	remindersInput sync.Map
	reminders      sync.Map
	reminderStatus sync.Map
}

func (r *ReminderRepository) PreHoldInfo(key string, data ReminderInfo) {
	r.remindersInput.Store(key, data)
}

func (r *ReminderRepository) PreLoad(key string) (ReminderInfo, error) {
	if v, ok := r.remindersInput.Load(key); ok {
		return v.(ReminderInfo), nil
	}
	return ReminderInfo{}, fmt.Errorf("not found")
}

func (r *ReminderRepository) HoldInfo(key string, data ReminderInfoExec) error {
	r.reminders.Store(key, data)
	if s := GetReminderStore(); s != nil {
		return s.UpsertDraft(context.Background(), key, data)
	}
	return nil
}

func (r *ReminderRepository) StoreInfo(key string, data ReminderInfoExec) error {
	s := GetReminderStore()
	if s == nil {
		r.reminderStatus.Store(key, data)
		r.reminders.Delete(key)
		return nil
	}
	if err := s.Confirm(context.Background(), key); err != nil {
		return err
	}
	r.reminders.Delete(key)
	return nil
}

func (r *ReminderRepository) Load(key string) (ReminderInfoExec, error) {
	if v, ok := r.reminders.Load(key); ok {
		return v.(ReminderInfoExec), nil
	}
	if s := GetReminderStore(); s != nil {
		return s.Get(context.Background(), key)
	}
	return ReminderInfoExec{}, fmt.Errorf("not found")
}

func (r *ReminderRepository) DeleteDraft(key string) error {
	r.reminders.Delete(key)
	s := GetReminderStore()
	if s == nil {
		return nil
	}
	return s.Delete(context.Background(), key)
}

type ReminderCmd struct{}

func NewReminderCmd() *ReminderCmd {
	repository.RemindChecker()
	return &ReminderCmd{}
}

func (n *ReminderCmd) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Printf("Entering ReminderCmd.Handle: Type=%s, InteractionID=%s, UserID=%s", i.Type, i.ID, i.Member.User.ID)
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		n.handleApplicationCommand(s, i)
	case discordgo.InteractionModalSubmit:
		n.handleModalSubmit(s, i)
	case discordgo.InteractionMessageComponent:
		n.handleMessageComponent(s, i)
	}
	log.Printf("Exiting ReminderCmd.Handle for user %s", i.Member.User.ID)
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
