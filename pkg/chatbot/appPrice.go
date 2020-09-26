package chatbot

import (
	"html/template"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/i18n"
)

type CommandAppPrice struct {
}

func (c CommandAppPrice) ID() string {
	return CAppPrice
}

func (CommandAppPrice) Regex() string {
	return `^[.|!]price\s?([a-zA-Z]{2})?\s(.*)`
}

func (CommandAppPrice) DisableCache() bool {
	return false
}

func (CommandAppPrice) PerProdCode() bool {
	return true
}

func (CommandAppPrice) Example() string {
	return ".price {region}? {game}"
}

func (CommandAppPrice) Description() template.HTML {

	var ccs []string
	for _, v := range i18n.GetProdCCs(true) {
		ccs = append(ccs, string(v.ProductCode))
	}

	//noinspection GoRedundantConversion
	return "Get the price of a game <small>(Allowed regions: " + template.HTML(strings.Join(ccs, ", ")) + ")</small>"
}

func (CommandAppPrice) Type() CommandType {
	return TypeGame
}

func (c CommandAppPrice) Output(msg *discordgo.MessageCreate, code steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	apps, err := elasticsearch.SearchAppsSimple(1, matches[2], []string{"id", "name", "prices"})
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + matches[2] + "** not found"
		return message, nil
	}

	app := apps[0]

	if matches[1] != "" {
		matches[1] = strings.ToLower(matches[1])
		if matches[1] == "gb" {
			matches[1] = "uk"
		}
		if steamapi.IsProductCC(matches[1]) {
			code = steamapi.ProductCC(matches[1])
		}
	}

	price := app.Prices.Get(code)

	if price.Exists {
		message.Content = app.GetName() + " is **" + price.GetFinal() + "** for " + strings.ToUpper(string(code))
		return message, nil
	}

	message.Content = app.GetName() + " has no price for " + strings.ToUpper(string(code))
	return message, nil
}
