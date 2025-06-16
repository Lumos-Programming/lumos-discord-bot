package cmd

import (
	"github.com/bwmarrin/discordgo"
	"log"
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
	log.Println("added command: ", i.Info().Name)
}

func (c *exec) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		name := i.ApplicationCommandData().Name
		if h, ok := c.cmds[name]; ok {
			h.Handle(s, i)
		} else {
			log.Printf("unknown command: %s", name)
		}
		return
	case discordgo.InteractionModalSubmit:
		customID := i.ModalSubmitData().CustomID
		if h, ok := c.modals[customID]; ok {
			h.Handle(s, i)
		} else {
			log.Printf("unknown modal: %s", customID)
		}
		return
	default:
		log.Printf("unhandled interaction, type: %s, id: %s", i.Type, i.ID)
	}
}
