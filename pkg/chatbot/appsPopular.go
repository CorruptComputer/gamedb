package chatbot

import (
	"fmt"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppsPopular struct {
}

func (c CommandAppsPopular) ID() string {
	return CAppsPopular
}

func (CommandAppsPopular) Regex() string {
	return `^[.|!](popular|top)`
}

func (CommandAppsPopular) DisableCache() bool {
	return false
}

func (CommandAppsPopular) PerProdCode() bool {
	return false
}

func (CommandAppsPopular) Example() string {
	return ".popular"
}

func (CommandAppsPopular) Description() string {
	return "Retrieve the most popular games this week"
}

func (CommandAppsPopular) Type() CommandType {
	return TypeGame
}

func (CommandAppsPopular) LegacyInputs(input string) map[string]string {
	return map[string]string{}
}

func (c CommandAppsPopular) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{}
}

func (CommandAppsPopular) Output(authorID string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Popular Games",
		URL:    config.C.GameDBDomain + "/games",
		Author: getAuthor(authorID),
		Color:  greenHexDec,
	}

	apps, err := mongo.PopularApps()
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	var code []string
	for k, v := range apps {

		if k == 0 {
			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: v.GetHeaderImage()}
		}

		code = append(code, fmt.Sprintf("%2d", k+1)+": "+humanize.Comma(int64(v.PlayerPeakWeek))+" - "+v.GetName())
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
