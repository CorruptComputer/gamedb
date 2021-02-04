package chatbot

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
)

type CommandPlayer struct {
}

func (c CommandPlayer) ID() string {
	return CPlayer
}

func (CommandPlayer) Regex() string {
	return `^[.|!](player|user)\s(.{2,32})`
}

func (CommandPlayer) DisableCache() bool {
	return false
}

func (CommandPlayer) PerProdCode() bool {
	return false
}

func (CommandPlayer) Example() string {
	return ".player {player}?"
}

func (CommandPlayer) Description() string {
	return "Retrieve information about a player"
}

func (CommandPlayer) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayer) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayer) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayer) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	var player Player

	if inputs["player"] == "" {

		provider, err := mysql.GetUserProviderByProviderID(oauth.ProviderDiscord, authorID)
		if err != nil {
			message.Content = "Please connect your Discord account first: <" + config.C.GameDBDomain + "/oauth/out/discord?page=settings>"
			return message, nil
		}

		provider, err = mysql.GetUserProviderByUserID(oauth.ProviderSteam, provider.UserID)
		if err != nil {
			message.Content = "Please connect your Steam account first: <" + config.C.GameDBDomain + "/oauth/out/steam?page=settings>"
			return message, nil
		}

		i, err := strconv.ParseInt(provider.ID, 10, 64)
		if err != nil || i == 0 {
			message.Content = "We had trouble finding your profile on Global Steam"
			return message, nil
		}

		player, err = mongo.GetPlayer(i)
		if err != nil {
			message.Content = "We had trouble finding your profile on Global Steam"
			return message, nil
		}

	} else {

		player, err = searchForPlayer(inputs["player"])
		if err == mongo.ErrNoDocuments {

			message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
			return message, nil

		} else if err != nil {
			return message, err
		}
	}

	var games string
	if player.GetGamesCount() == 0 {
		games = "Profile set to private"
	} else {
		games = humanize.Comma(int64(player.GetGamesCount())) + " (" + helpers.OrdinalComma(player.GetRanks()[string(mongo.RankKeyGames)]) + ")"
	}

	var playtime string
	if player.GetPlaytime() == 0 {
		playtime = "Profile set to private"
	} else {
		playtime = helpers.GetTimeLong(player.GetPlaytime(), 3) + " (" + helpers.OrdinalComma(player.GetRanks()[string(mongo.RankKeyPlaytime)]) + ")"
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:     player.GetName(),
		URL:       config.C.GameDBDomain + player.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
		Footer:    getFooter(),
		Color:     greenHexDec,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Level",
				Value: humanize.Comma(int64(player.GetLevel())) + " (" + helpers.OrdinalComma(player.GetRanks()[string(mongo.RankKeyLevel)]) + ")",
			},
			{
				Name:  "Games",
				Value: games,
			},
			{
				Name:  "Achievements",
				Value: humanize.Comma(int64(player.GetAchievements())) + " (" + helpers.OrdinalComma(player.GetRanks()[string(mongo.RankKeyAchievements)]) + ")",
			},
			{
				Name:  "Badges",
				Value: humanize.Comma(int64(player.GetBadges())) + " (" + helpers.OrdinalComma(player.GetRanks()[string(mongo.RankKeyBadges)]) + ")",
			},
			{
				Name:  "Foil Badges",
				Value: humanize.Comma(int64(player.GetBadgesFoil())) + " (" + helpers.OrdinalComma(player.GetRanks()[string(mongo.RankKeyBadgesFoil)]) + ")",
			},
			{
				Name:  "Playtime",
				Value: playtime,
			},
		},
	}

	return message, nil
}
