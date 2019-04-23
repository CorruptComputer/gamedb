package chatbot

import (
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/mongo"
)

type CommandPlayerLevel struct {
}

func (CommandPlayerLevel) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.level (.*)")
}

func (c CommandPlayerLevel) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[1], mongo.M{"_id": 1, "persona_name": 1, "level": 1})
	if err != nil {
		return message, err
	}

	message.Content = player.GetName() + " is level **" + strconv.Itoa(player.Level) + "**"

	return message, nil
}

func (CommandPlayerLevel) Example() string {
	return ".level {player_name}"
}

func (CommandPlayerLevel) Description() string {
	return "Get the level of a player"
}
