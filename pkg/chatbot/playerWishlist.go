package chatbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerWishlist struct {
}

func (c CommandPlayerWishlist) ID() string {
	return CPlayerWishlist
}

func (CommandPlayerWishlist) Regex() string {
	return `^[.|!]wishlist (.+)`
}

func (CommandPlayerWishlist) DisableCache() bool {
	return false
}

func (CommandPlayerWishlist) PerProdCode() bool {
	return false
}

func (CommandPlayerWishlist) AllowDM() bool {
	return false
}

func (CommandPlayerWishlist) Example() string {
	return ".wishlist {player}"
}

func (CommandPlayerWishlist) Description() string {
	return "Retrieve a player's wishlist"
}

func (CommandPlayerWishlist) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerWishlist) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[1],
	}
}

func (c CommandPlayerWishlist) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandPlayerWishlist) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

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

	wishlistApps, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 10, bson.D{{"order", 1}}, nil)
	if err != nil {
		return message, err
	}

	if len(wishlistApps) > 10 {
		wishlistApps = wishlistApps[0:10]
	}

	if len(wishlistApps) > 0 {

		var code []string
		for _, app := range wishlistApps {

			var rank string
			if app.Order > 0 {
				rank = strconv.Itoa(app.Order)
			} else {
				rank = "*"
			}

			code = append(code, fmt.Sprintf("%2s", rank)+": "+app.GetName())
		}

		message.Embed = &discordgo.MessageEmbed{
			Title:       "Wishlist Items",
			URL:         player.GetPathAbsolute() + "#wishlist",
			Author:      getAuthor(authorID),
			Color:       greenHexDec,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Description: "```" + strings.Join(code, "\n") + "```",
		}

	} else {
		message.Content = player.GetName() + " has no wishlist items, or a profile set to private"
	}

	return message, nil
}
