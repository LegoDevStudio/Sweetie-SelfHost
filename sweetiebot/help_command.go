package sweetiebot

import (
  "strings"
)

type HelpCommand struct {
}

func (c *HelpCommand) Name() string {
  return "Help";  
}
func (c *HelpCommand) Process(args []string) string {
  if len(args) == 0 {
    s := []string{"Sweetie Bot knows the following commands. For more information on a specific command, type !help [command].\n"}
    for k, v := range sb.commands {
      s = append(s, k + ": " + v.c.UsageShort())
    }
    return "```" + strings.Join(s, "\n") + "```"
  }
  v, ok := sb.commands[strings.ToLower(args[0])]
  if !ok {
    return "``` Sweetie Bot doesn't recognize the '" + args[0] + "' command. You can check what commands Sweetie Bot knows by typing !help.```"
  }
  return "```> !" + args[0] + " " + v.c.Usage() + "```"
}
func (c *HelpCommand) Usage() string { 
  return FormatUsage(c, "[command]", "Lists all available commands Sweetie Bot knows, or gives information about the given command. Of course, you should have figured this out by now, since you just typed !help help for some reason.") 
}
func (c *HelpCommand) UsageShort() string { return "Generates the list you are looking at right now." }
func (c *HelpCommand) Roles() []string { return []string{} }

type AboutCommand struct {
}

func (c *AboutCommand) Name() string {
  return "About";  
}
func (c *AboutCommand) Process(args []string) string {
  s := "```Sweetie Bot version " + sb.version
  if sb.config.Debug {
    return s + " [debug]```"
  } 
  return s + " [release]```"
}
func (c *AboutCommand) Usage() string { 
  return FormatUsage(c, "", "Display information about Sweetie Bot. What, did you think it would do something else?") 
}
func (c *AboutCommand) UsageShort() string { return "Displays information about Sweetie Bot." }
func (c *AboutCommand) Roles() []string { return []string{} }