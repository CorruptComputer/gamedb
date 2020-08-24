package chatbot

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerRecent struct {
}

func (c CommandPlayerRecent) ID() string {
	return CPlayerRecent
}

func (CommandPlayerRecent) Regex() string {
	return `^[.|!]recent (.{2,32})$`
}

func (CommandPlayerRecent) DisableCache() bool {
	return false
}

func (CommandPlayerRecent) Example() string {
	return ".recent {player}"
}

func (CommandPlayerRecent) Description() template.HTML {
	return "Returns the last 10 games played by user"
}

func (CommandPlayerRecent) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerRecent) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	player, q, err := mongo.SearchPlayer(matches[1], nil)
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	recent, err := mongo.GetRecentApps(player.ID, 0, 10, bson.D{{"playtime_2_weeks", -1}})
	if err != nil {
		return message, err
	}

	if len(recent) > 10 {
		recent = recent[0:10]
	}

	if len(recent) > 0 {

		message.Content = "<@" + msg.Author.ID + ">"
		message.Embed = &discordgo.MessageEmbed{
			Title:  "Recent Games",
			URL:    config.C.GameDBDomain + player.GetPath() + "#games",
			Author: getAuthor(msg.Author.ID),
		}

		var code []string

		for k, app := range recent {

			if k == 0 {

				avatar := app.GetIcon()
				if strings.HasPrefix(avatar, "/") {
					avatar = "https://gamedb.online" + avatar
				}

				message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: avatar}
			}

			code = append(code, fmt.Sprintf("%2d", k+1)+": "+app.AppName+" - "+helpers.GetTimeShort(app.PlayTime2Weeks, 2))
		}

		message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
