package cmd

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type exec struct {
	cmds     map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
	cmdInfos map[string]*discordgo.ApplicationCommand
}

func (c *exec) Add(app *discordgo.ApplicationCommand, h func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	c.cmds[app.Name] = h
	c.cmdInfos[app.Name] = app
}

func (c *exec) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := c.cmds[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	} else {
		log.Printf("unknown command: %s", i.ApplicationCommandData().Name)
	}
}
