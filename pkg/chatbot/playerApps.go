package chatbot

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerApps struct {
}

func (c CommandPlayerApps) ID() string {
	return CPlayerApps
}

func (CommandPlayerApps) Regex() string {
	return `^[.|!](games|apps) (.{2,32})`
}

func (CommandPlayerApps) DisableCache() bool {
	return false
}

func (CommandPlayerApps) PerProdCode() bool {
	return false
}

func (CommandPlayerApps) Example() string {
	return ".games {player}"
}

func (CommandPlayerApps) Description() string {
	return "Retrieve the number of games in a player's library"
}

func (CommandPlayerApps) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerApps) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayerApps) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerApps) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	player, q, err := mongo.SearchPlayer(inputs["player"], bson.M{"_id": 1, "persona_name": 1, "games_count": 1, "ranks": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.apps")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Sucess response
	var rank = "Unranked"
	if val, ok := player.Ranks[string(mongo.RankKeyGames)]; ok {
		rank = helpers.OrdinalComma(val)
	}

	if player.GamesCount > 0 {
		message.Embed = &discordgo.MessageEmbed{
			Title:     player.GetName(),
			URL:       config.C.GameDBDomain + player.GetPath(),
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Footer:    getFooter(),
			Color:     greenHexDec,
			Image:     &discordgo.MessageEmbedImage{URL: charts.GetPlayerChart(c.ID(), player.ID, influxHelper.InfPlayersGames, "Games")},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Games",
					Value:  strconv.Itoa(player.GamesCount),
					Inline: true,
				},
				{
					Name:   "Rank",
					Value:  rank,
					Inline: true,
				},
			},
		}
	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
