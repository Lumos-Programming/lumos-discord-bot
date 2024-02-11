package cmd

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type cmd struct {
	e exec
	d *func()
}

type SubCmd interface {
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
	Info() *discordgo.ApplicationCommand
}

func (c exec) Activate(s *discordgo.Session) cmd {
	for _, v := range c.cmds {
		if v.Info() == nil {
			continue
		}
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v.Info())
		if err != nil {
			log.Printf("failed to create command %s, err: %v", v.Info().Name, err)
		}
	}
	d := s.AddHandler(c.Handle)
	return cmd{e: c, d: &d}
}

func (c *cmd) Deactivate() {
	if c.d == nil {
		return
	}
	(*c.d)()
	c.d = nil
}
