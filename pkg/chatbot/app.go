package chatbot

import (
	"regexp"

	"github.com/Jleagle/steam-go/steam"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/sql"
)

type CommandApp struct {
}

func (CommandApp) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.(app|game) (.*)")
}

func (c CommandApp) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	app, err := sql.SearchApp(matches[2], nil)
	if err != nil {
		return message, err
	}

	price, err := app.GetPrice(steam.CountryUS)
	if err != nil {
		// log.Err(err)
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:  app.GetName(),
		URL:    "https://gamedb.online" + app.GetPath(),
		Author: author,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: app.GetHeaderImage(),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Release Date",
				Value: app.GetReleaseDateNice(),
			},
			{
				Name:  "Price",
				Value: price.GetFinal(),
			},
			{
				Name:  "Review Score",
				Value: app.GetReviewScore(),
			},
		},
	}

	return message, nil
}

func (CommandApp) Example() string {
	return ".game {game_name}"
}

func (CommandApp) Description() string {
	return "Get info on a game"
}

func (CommandApp) Type() CommandType {
	return TypeGame
}
