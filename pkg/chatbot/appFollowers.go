package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppFollowers struct {
}

func (c CommandAppFollowers) ID() string {
	return CAppFollowers
}

func (CommandAppFollowers) Regex() string {
	return `^[.|!]followers (.*)`
}

func (CommandAppFollowers) DisableCache() bool {
	return false
}

func (CommandAppFollowers) PerProdCode() bool {
	return false
}

func (CommandAppFollowers) Example() string {
	return ".followers {game}"
}

func (CommandAppFollowers) Description() string {
	return "Gets the number of followers for a game."
}

func (CommandAppFollowers) Type() CommandType {
	return TypeGame
}

func (c CommandAppFollowers) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"game": matches[1],
	}
}

func (c CommandAppFollowers) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "game",
			Description: "The name or ID of the game",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandAppFollowers) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	apps, err := elasticsearch.SearchAppsSimple(1, inputs["game"])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + inputs["game"] + "** not found on Steam"
		return message, nil
	}

	app, err := mongo.GetApp(apps[0].ID)
	if err != nil {
		return message, err
	}

	if app.GroupID == "" {
		message.Content = app.GetName() + " has no followers"
		return
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:       app.GetName(),
		Description: humanize.Comma(int64(app.GroupFollowers)) + " followers",
		URL:         config.C.GameDBDomain + app.GetPath(),
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage()},
		Footer:      getFooter(),
		Color:       2664261,
		Image: &discordgo.MessageEmbedImage{
			URL: charts.GetGroupChart(c.ID(), app.GroupID),
		},
	}

	return message, nil
}
