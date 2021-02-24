package chatbot

import (
	"fmt"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppsTrending struct {
}

func (c CommandAppsTrending) ID() string {
	return CAppsTrending
}

func (CommandAppsTrending) Regex() string {
	return `^[.|!]trending$`
}

func (CommandAppsTrending) DisableCache() bool {
	return false
}

func (CommandAppsTrending) PerProdCode() bool {
	return false
}

func (CommandAppsTrending) Example() string {
	return ".trending"
}

func (CommandAppsTrending) Description() string {
	return "Retrieve the most trending games"
}

func (CommandAppsTrending) Type() CommandType {
	return TypeGame
}

func (CommandAppsTrending) LegacyInputs(_ string) map[string]string {
	return map[string]string{}
}

func (c CommandAppsTrending) Slash() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (CommandAppsTrending) Output(authorID string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Games",
		URL:    config.C.GameDBDomain + "/games/trending",
		Author: getAuthor(authorID),
		Color:  greenHexDec,
	}

	apps, err := mongo.TrendingApps()
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	var code []string
	for k, app := range apps {

		if k == 0 {
			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage(), Width: 460, Height: 215}
		}

		code = append(code, fmt.Sprintf("%2d", k+1)+": "+app.GetTrend()+" "+app.GetName())
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
