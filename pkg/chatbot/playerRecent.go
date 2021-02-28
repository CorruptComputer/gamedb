package chatbot

import (
	"fmt"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerRecent struct {
}

func (c CommandPlayerRecent) ID() string {
	return CPlayerRecent
}

func (CommandPlayerRecent) Regex() string {
	return `^[.|!]recent (.+)`
}

func (CommandPlayerRecent) DisableCache() bool {
	return false
}

func (CommandPlayerRecent) PerProdCode() bool {
	return false
}

func (CommandPlayerRecent) AllowDM() bool {
	return false
}

func (CommandPlayerRecent) Example() string {
	return ".recent {player}"
}

func (CommandPlayerRecent) Description() string {
	return "Retrieve a player's last opened games"
}

func (CommandPlayerRecent) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerRecent) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[1],
	}
}

func (c CommandPlayerRecent) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandPlayerRecent) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	if inputs["player"] == "" {
		message.Content = "Missing player name"
		return message, nil
	}

	player, err := searchForPlayer(inputs["player"])
	if err == elasticsearch.ErrNoResult || err == steamapi.ErrProfileMissing {

		message.Content = "Player **" + inputs["player"] + "** not found, they may be set to private, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	recent, err := mongo.GetRecentApps(player.ID, 0, 10, bson.D{{"playtime_2_weeks", -1}})
	if err != nil {
		return message, err
	}

	if len(recent) > 10 {
		recent = recent[0:10]
	}

	if len(recent) > 0 {

		var code []string
		for k, app := range recent {
			code = append(code, fmt.Sprintf("%2d", k+1)+": "+helpers.GetTimeShort(app.PlayTime2Weeks, 2)+" - "+app.AppName)
		}

		message.Embed = &discordgo.MessageEmbed{
			Title:       "Recent Games",
			URL:         player.GetPathAbsolute() + "#games",
			Author:      getAuthor(authorID),
			Color:       greenHexDec,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Description: "```" + strings.Join(code, "\n") + "```",
		}

	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
