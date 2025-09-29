package reminder

import (
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

var reminders sync.Map      // map[string]ReminderInfo for temporary storage
var reminderStatus sync.Map // map[string]bool for execute

type ReminderCmd struct{}

func NewReminderCmd() *ReminderCmd {
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

//Readyイベントの受け取り方が分からない
//func (n *ReminderCmd) readyHandler(i *discordgo.Ready) {
//	log.Printf("ReminderCmd.HandleReady for user %s", i.User.ID)
//	n.RemindChecker()
//}
