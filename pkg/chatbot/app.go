package chatbot

import (
	"html/template"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandApp struct {
}

func (c CommandApp) ID() string {
	return CApp
}

func (CommandApp) Regex() string {
	return `^[.|!](app|game) (.*)`
}

func (CommandApp) DisableCache() bool {
	return false
}

func (CommandApp) Example() string {
	return ".game {game}"
}

func (CommandApp) Description() template.HTML {
	return "Get info on a game"
}

func (CommandApp) Type() CommandType {
	return TypeGame
}

func (c CommandApp) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	apps, err := elasticsearch.SearchAppsSimple(1, matches[2])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + matches[2] + "** not found"
		return message, nil
	}

	app, err := mongo.GetApp(apps[0].ID)
	if err != nil {
		return message, err
	}

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = getAppEmbed(app)

	img, err := charts.GetAppChart(app)
	if err != nil {
		log.Err(err)
	} else {
		message.Files = append(message.Files, &discordgo.File{
			Name:        "app-" + strconv.Itoa(app.ID) + ".png",
			ContentType: "image/png",
			Reader:      img,
		})
	}

	return message, nil
}
