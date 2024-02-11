package cmd

import "github.com/bwmarrin/discordgo"

type cmd struct {
	e exec
	d *func()
}

func (c exec) Activate(s *discordgo.Session) cmd {
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
