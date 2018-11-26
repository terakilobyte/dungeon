package commands

import (
	"strings"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"

	twitch "github.com/gempir/go-twitch-irc"
)

// Command struct
type Command struct {
	client  *twitch.Client
	channel string
}

// RawCommand struct
type RawCommand struct {
	Collection *mongo.Collection
	Args       string
	Channel    string
	User       twitch.User
}

var (
	commandMap        map[string]fn
	project           string
	admins            = []string{"swarmlogic", "aidenmontgomery", "swarmlogic_bot"}
	badgesWeCareAbout = []string{"broadcaster", "moderator"}
	pollInProgress    = false
	currentPoll       *poll
	startTime         time.Time
	channel           string
)

// NewCommand returns an instantiated Command struct
func NewCommand(client *twitch.Client, twitchChannel string) *Command {
	project = "programming"
	startTime = time.Now()
	channel = twitchChannel
	commandMap = map[string]fn{
		"time":       getTime,
		"project":    getProject,
		"setproject": setProject,
		"8ball":      eightball,
		"uptime":     uptime,
		"poll":       makePoll,
		"vote":       votePoll,
		"options":    optionsPoll,
		"github":     github,
		"commands":   getCommands,
		"mytime":     myTime,
	}
	return &Command{client, channel}
}

// HandleCommand handles...
func (c *Command) HandleCommand(rc RawCommand) {
	if rc.User.Username == "swarmlogic_bot" {
		return
	}
	parsed := strings.Split(rc.Args, " ")
	cArgs := &commandArgs{parsed[1:], rc.User, c.client, rc.Collection}
	for k, v := range commandMap {
		if k == parsed[0] {
			c.client.Say(channel, v(cArgs))
			return
		}
	}
	c.client.Whisper(rc.User.DisplayName, "that isn't a valid command")
}
