package tryhackme_achievement

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	cmdName    = "tryhackme-share"
	modalID    = "tryhackme-share"
	titleID    = "challenge_title"
	urlID      = "challenge_url"
	commentsID = "comments"
)

// TryHackMeCmd encapsulates all logic for the `/share` command.
type TryHackMeCmd struct {
	GuildID string
}

// NewTryHackMeCmd returns a new command handler.
func NewTryHackMeCmd(guildID string) *TryHackMeCmd {
	return &TryHackMeCmd{GuildID: guildID}
}

// RegisterCommand registers/updates the slash-command every time the bot becomes READY.
func (c *TryHackMeCmd) RegisterCommand(s *discordgo.Session) error {
	_, err := s.ApplicationCommandCreate(
		s.State.User.ID,
		c.GuildID,
		&discordgo.ApplicationCommand{
			Name:        cmdName,
			Description: "Share your TryHackMe achievement!",
		},
	)
	return err
}

// Handle must be added with dg.AddHandler.
// It processes both the slash-command invocation and the modal submit event.
func (c *TryHackMeCmd) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name != cmdName {
			return
		}

		// Send modal
		modal := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: modalID,
				Title:    "TryHackMe達成共有しよう!",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    titleID,
								Label:       "Room名",
								Style:       discordgo.TextInputShort,
								Placeholder: "e.g. 'Blue Team Basics'",
								Required:    true,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    urlID,
								Label:       "RoomのURL",
								Style:       discordgo.TextInputShort,
								Placeholder: "https://tryhackme.com/room/…",
								Required:    true,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: commentsID,
								Label:    "Comments (任意)",
								Style:    discordgo.TextInputParagraph,
								Required: false,
							},
						},
					},
				},
			},
		}

		if err := s.InteractionRespond(i.Interaction, modal); err != nil {
			log.Printf("tryhackme: modal respond error: %v", err)
		}

	case discordgo.InteractionModalSubmit:
		if i.ModalSubmitData().CustomID != modalID {
			return
		}

		// Extract inputs
		var title, url, comment string
		for _, cmp := range i.ModalSubmitData().Components {
			row := cmp.(*discordgo.ActionsRow)
			for _, inner := range row.Components {
				input := inner.(*discordgo.TextInput)
				switch input.CustomID {
				case titleID:
					title = input.Value
				case urlID:
					url = input.Value
				case commentsID:
					comment = input.Value
				}
			}
		}

		name := i.Member.Nick
		if name == "" {
			name = i.Member.User.GlobalName
		}
		embed := &discordgo.MessageEmbed{
			Title: "🎖TryHackMe達成共有⭐️",
			// orange
			Color:     0xFFA500,
			Timestamp: time.Now().Format(time.RFC3339),
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Room名", Value: title, Inline: true},
				{Name: "リンク", Value: url, Inline: true},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    fmt.Sprintf("%s(%s)", i.Member.Nick, i.Member.User.Username),
				IconURL: i.Member.User.AvatarURL(""),
			},
		}
		if comment != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Comments",
				Value: comment,
			})
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		}); err != nil {
			log.Printf("tryhackme: embed respond error: %v", err)
		}
	}
}

func (c *TryHackMeCmd) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmdName,
		Description: "Share your achievement on TryHackMe!",
	}
}

func (c *TryHackMeCmd) ModalCustomIDs() []string {
	return []string{modalID}
}
