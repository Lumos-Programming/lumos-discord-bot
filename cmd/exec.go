package cmd

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type exec struct {
	cmds map[string]SubCmd
}

func NewExec() *exec {
	return &exec{}
}

func (c *exec) Add(i SubCmd) {
	if c.cmds == nil {
		c.cmds = make(map[string]SubCmd)
	}
	c.cmds[i.Info().Name] = i
}

func (c *exec) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := c.cmds[i.ApplicationCommandData().Name]; ok {
		h.Handle(s, i)
	} else {
		log.Printf("unknown command: %s", i.ApplicationCommandData().Name)
	}
}
