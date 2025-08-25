package cmd

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

type exec struct {
	cmds   map[string]SubCmd
	modals map[string]SubCmd
}

func NewExec() *exec {
	return &exec{
		cmds:   make(map[string]SubCmd),
		modals: make(map[string]SubCmd),
	}
}

func (c *exec) Add(i SubCmd) {
	c.cmds[i.Info().Name] = i
	for _, cid := range i.ModalCustomIDs() {
		c.modals[cid] = i
	}
	log.Println("Added command: ", i.Info().Name)
}

func (c *exec) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		name := i.ApplicationCommandData().Name
		if h, ok := c.cmds[name]; ok {
			h.Handle(s, i)
		} else {
			log.Printf("Unknown command: %s", name)
		}
	case discordgo.InteractionModalSubmit:
		customID := i.ModalSubmitData().CustomID
		if h, ok := c.modals[customID]; ok {
			h.Handle(s, i)
		} else {
			log.Printf("Unknown modal: %s", customID)
		}
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		for name, h := range c.cmds {
			if strings.HasPrefix(customID, name+"-") {
				log.Printf("Routing component to command: %s", name)
				h.Handle(s, i)
				return
			}
		}
		log.Printf("Unknown component: %s", customID)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "エラー：不明なボタン操作です。",
			},
		}); err != nil {
			log.Printf("Failed to respond to unknown component: %v", err)
		}
	default:
		log.Printf("Unhandled interaction type: %s, ID: %s", i.Type, i.ID)
	}
}
