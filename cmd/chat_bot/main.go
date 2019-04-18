package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/chat_bot"
	"github.com/gamedb/website/pkg/config"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
)

const debugAuthorID = "145456943912189952"

func main() {

	if !config.Config.IsProd() && !config.Config.IsLocal() {
		log.Err("Prod & local only")
	}

	discord, err := discordgo.New("Bot " + config.Config.DiscordBotToken.Get())
	if err != nil {
		fmt.Println(err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		if config.Config.IsLocal() && m.Author.ID != debugAuthorID {
			return
		}

		// Don't reply to bots
		if m.Author.Bot {
			return
		}

		for _, v := range chat_bot.CommandRegister {

			if v.Regex().MatchString(m.Message.Content) {

				private, err := isPrivateChannel(s, m)
				if err != nil {
					fmt.Println(err)
					return
				}

				chanID := m.ChannelID

				if private {

					st, err := s.UserChannelCreate(m.Author.ID)
					if err != nil {
						fmt.Println(err)
						return
					}

					chanID = st.ID
				}

				_, err = s.ChannelMessageSend(chanID, v.Output(m.Message.Content))
				if err != nil {
					fmt.Println(err)
					return
				}

				return
			}
		}
	})

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	helpers.KeepAlive()
}

func isPrivateChannel(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}
