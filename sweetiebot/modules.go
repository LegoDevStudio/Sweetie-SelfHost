package sweetiebot

import (
	"fmt"
	"strings"

	"github.com/blackhole12/discordgo"
)

// Giving each possible hook function its own interface ensures each module
// only has to define functions for the hooks it actually cares about

type ModuleOnEvent interface {
	Module
	OnEvent(*GuildInfo, *discordgo.Event)
}

type ModuleOnMessageCreate interface {
	Module
	OnMessageCreate(*GuildInfo, *discordgo.Message)
}

type ModuleOnMessageUpdate interface {
	Module
	OnMessageUpdate(*GuildInfo, *discordgo.Message)
}

type ModuleOnMessageDelete interface {
	Module
	OnMessageDelete(*GuildInfo, *discordgo.Message)
}

type ModuleOnPresenceUpdate interface {
	Module
	OnPresenceUpdate(*GuildInfo, *discordgo.PresenceUpdate)
}

type ModuleOnVoiceStateUpdate interface {
	Module
	OnVoiceStateUpdate(*GuildInfo, *discordgo.VoiceState)
}

type ModuleOnGuildUpdate interface {
	Module
	OnGuildUpdate(*GuildInfo, *discordgo.Guild)
}

type ModuleOnGuildMemberAdd interface {
	Module
	OnGuildMemberAdd(*GuildInfo, *discordgo.Member)
}

type ModuleOnGuildMemberRemove interface {
	Module
	OnGuildMemberRemove(*GuildInfo, *discordgo.Member)
}

type ModuleOnGuildMemberUpdate interface {
	Module
	OnGuildMemberUpdate(*GuildInfo, *discordgo.Member)
}

type ModuleOnGuildBanAdd interface {
	Module
	OnGuildBanAdd(*GuildInfo, *discordgo.GuildBanAdd)
}

type ModuleOnGuildBanRemove interface {
	Module
	OnGuildBanRemove(*GuildInfo, *discordgo.GuildBanRemove)
}

type ModuleOnGuildRoleDelete interface {
	Module
	OnGuildRoleDelete(*GuildInfo, *discordgo.GuildRoleDelete)
}

type ModuleOnCommand interface {
	Module
	OnCommand(*GuildInfo, *discordgo.Message) bool
}

type ModuleOnIdle interface {
	Module
	OnIdle(*GuildInfo, *discordgo.Channel)
	IdlePeriod(*GuildInfo) int64
}

type ModuleOnTick interface {
	Module
	OnTick(*GuildInfo)
}

// Modules monitor all incoming messages and users that have joined a given channel.
type Module interface {
	Name() string
	Commands() []Command
	Description() string
}

type CommandUsageParam struct {
	Name     string
	Desc     string
	Optional bool
	Variadic bool
}
type CommandUsage struct {
	Desc   string
	Params []CommandUsageParam
}

// Commands are any command that is addressed to the bot, optionally restricted by role.
type Command interface {
	Name() string
	Process([]string, *discordgo.Message, []int, *GuildInfo) (string, bool, *discordgo.MessageEmbed)
	Usage(*GuildInfo) *CommandUsage
	UsageShort() string
}

func (info *GuildInfo) IsModuleDisabled(name string) string {
	_, ok := info.config.Modules.Disabled[strings.ToLower(name)]
	if ok {
		return " [disabled]"
	}
	return ""
}

func (info *GuildInfo) IsCommandDisabled(name string) string {
	str := ""
	_, disabled := info.config.Modules.CommandDisabled[strings.ToLower(name)]
	_, restricted := sb.RestrictedCommands[strings.ToLower(name)]
	if restricted && !sb.IsDBGuild(info) {
		str += " [not available]"
	} else if disabled {
		str += " [disabled]"
	}
	return str
}

func (info *GuildInfo) GetRoles(c Command) string {
	m, ok := info.config.Modules.CommandRoles[strings.ToLower(c.Name())]
	if !ok {
		return ""
	}

	sb.dg.State.RLock()
	defer sb.dg.State.RUnlock()
	_, reverse := m["!"]
	s := make([]string, 0, len(m))
	for k, _ := range m {
		r, err := sb.dg.State.Role(info.ID, k)
		if err == nil {
			s = append(s, r.Name)
		}
	}

	if reverse {
		return "Any role except " + strings.Join(s, ", ")
	} else {
		return strings.Join(s, ", ")
	}
}

func (info *GuildInfo) GetChannels(c Command) string {
	m, ok := info.config.Modules.CommandChannels[strings.ToLower(c.Name())]
	if !ok {
		return ""
	}

	sb.dg.State.RLock()
	defer sb.dg.State.RUnlock()
	s := make([]string, 0, len(m))
	for k, _ := range m {
		c, err := sb.dg.State.Channel(k)
		if err == nil {
			s = append(s, "#"+c.Name)
		}
	}

	return strings.Join(s, ", ")
}

func (info *GuildInfo) FormatUsage(c Command, usage *CommandUsage) *discordgo.MessageEmbed {
	r := info.GetRoles(c)
	ch := info.GetChannels(c)
	fields := make([]*discordgo.MessageEmbedField, 0, len(usage.Params))
	use := "> " + info.config.Basic.CommandPrefix + strings.ToLower(c.Name())
	for _, v := range usage.Params {
		opt := ""
		if v.Optional {
			opt = " [OPTIONAL]"
			use += fmt.Sprintf(" [%s]", v.Name)
		} else {
			use += fmt.Sprintf(" {%s}", v.Name)
		}
		if v.Variadic {
			opt = " (...) " + opt
			use += "..."
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: v.Name + opt, Value: v.Desc, Inline: false})
	}

	if len(ch) > 0 {
		ch = fmt.Sprintf("Available on: %s", ch)
	}
	embed := &discordgo.MessageEmbed{
		Type: "rich",
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://github.com/blackhole12/sweetiebot#configuration",
			Name:    c.Name() + " Command",
			IconURL: fmt.Sprintf("https://cdn.discordapp.com/avatars/%v/%s.jpg", sb.SelfID, sb.SelfAvatar),
		},
		Color:       0xaaaaaa,
		Description: fmt.Sprintf("```%s```\n%s\n\n%s", use, usage.Desc, ch),
		Fields:      fields,
	}

	if len(r) > 0 {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "Only usable by: " + r}
	}
	return embed
}

type ModuleHooks struct {
	OnEvent             []ModuleOnEvent
	OnMessageCreate     []ModuleOnMessageCreate
	OnMessageUpdate     []ModuleOnMessageUpdate
	OnMessageDelete     []ModuleOnMessageDelete
	OnPresenceUpdate    []ModuleOnPresenceUpdate
	OnGuildUpdate       []ModuleOnGuildUpdate
	OnGuildMemberAdd    []ModuleOnGuildMemberAdd
	OnGuildMemberRemove []ModuleOnGuildMemberRemove
	OnGuildMemberUpdate []ModuleOnGuildMemberUpdate
	OnGuildBanAdd       []ModuleOnGuildBanAdd
	OnGuildBanRemove    []ModuleOnGuildBanRemove
	OnGuildRoleDelete   []ModuleOnGuildRoleDelete
	OnCommand           []ModuleOnCommand
	OnIdle              []ModuleOnIdle
	OnTick              []ModuleOnTick
}

func (info *GuildInfo) RegisterModule(m Module) {
	if h, ok := m.(ModuleOnEvent); ok {
		info.hooks.OnEvent = append(info.hooks.OnEvent, h)
	}
	if h, ok := m.(ModuleOnMessageCreate); ok {
		info.hooks.OnMessageCreate = append(info.hooks.OnMessageCreate, h)
	}
	if h, ok := m.(ModuleOnMessageUpdate); ok {
		info.hooks.OnMessageUpdate = append(info.hooks.OnMessageUpdate, h)
	}
	if h, ok := m.(ModuleOnMessageDelete); ok {
		info.hooks.OnMessageDelete = append(info.hooks.OnMessageDelete, h)
	}
	if h, ok := m.(ModuleOnPresenceUpdate); ok {
		info.hooks.OnPresenceUpdate = append(info.hooks.OnPresenceUpdate, h)
	}
	if h, ok := m.(ModuleOnGuildUpdate); ok {
		info.hooks.OnGuildUpdate = append(info.hooks.OnGuildUpdate, h)
	}
	if h, ok := m.(ModuleOnGuildMemberAdd); ok {
		info.hooks.OnGuildMemberAdd = append(info.hooks.OnGuildMemberAdd, h)
	}
	if h, ok := m.(ModuleOnGuildMemberRemove); ok {
		info.hooks.OnGuildMemberRemove = append(info.hooks.OnGuildMemberRemove, h)
	}
	if h, ok := m.(ModuleOnGuildMemberUpdate); ok {
		info.hooks.OnGuildMemberUpdate = append(info.hooks.OnGuildMemberUpdate, h)
	}
	if h, ok := m.(ModuleOnGuildBanAdd); ok {
		info.hooks.OnGuildBanAdd = append(info.hooks.OnGuildBanAdd, h)
	}
	if h, ok := m.(ModuleOnGuildBanRemove); ok {
		info.hooks.OnGuildBanRemove = append(info.hooks.OnGuildBanRemove, h)
	}
	if h, ok := m.(ModuleOnGuildRoleDelete); ok {
		info.hooks.OnGuildRoleDelete = append(info.hooks.OnGuildRoleDelete, h)
	}
	if h, ok := m.(ModuleOnCommand); ok {
		info.hooks.OnCommand = append(info.hooks.OnCommand, h)
	}
	if h, ok := m.(ModuleOnIdle); ok {
		fmt.Println("OnIdle")
		info.hooks.OnIdle = append(info.hooks.OnIdle, h)
	}
	if h, ok := m.(ModuleOnTick); ok {
		info.hooks.OnTick = append(info.hooks.OnTick, h)
	}
}
