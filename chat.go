package main

import (
	"net/http"
	"os"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/websockets"
)

const (
	guildID          = "407493776597057538"
	generalChannelID = "407493777058693121"
)

var (
	discordSession *discordgo.Session
)

func init() {

	var err error

	// Get client
	discordSession, err = discordgo.New("Bot " + os.Getenv("STEAM_DISCORD_BOT_TOKEN"))
	if err != nil {
		logger.Error(err)
	}

	// Add websocket listener
	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if !m.Author.Bot {
			websockets.Send(websockets.CHAT, chatPayload{
				AuthorID:     m.Author.ID,
				AuthorUser:   m.Author.Username,
				AuthorAvatar: m.Author.Avatar,
				Content:      m.Content,
			})
		}
	})

	// Open connection
	err = discordSession.Open()
	if err != nil {
		logger.Error(err)
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {

	// Get ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		id = generalChannelID
	}

	// Get channels
	channelsResponse, err := discordSession.GuildChannels(guildID)
	if err != nil {
		logger.Error(err)
	}

	channels := make([]*discordgo.Channel, 0)
	for _, v := range channelsResponse {
		if v.Type == discordgo.ChannelTypeGuildText {
			channels = append(channels, v)
		}
	}

	// Get messages
	messagesResponse, err := discordSession.ChannelMessages(id, 10, "", "", "")
	if err != nil {
		logger.Error(err)
	}

	messages := make([]*discordgo.Message, 0)
	for _, v := range messagesResponse {
		if !v.Author.Bot {
			messages = append(messages, v)
		}
	}

	// Template
	template := chatTemplate{}
	template.Channels = channels
	template.Messages = messages
	template.ChannelID = id

	returnTemplate(w, "chat", template)
}

type chatTemplate struct {
	GlobalTemplate
	Channels  []*discordgo.Channel
	Messages  []*discordgo.Message
	ChannelID string
}

type chatPayload struct {
	AuthorID     string `json:"author_id"`
	AuthorUser   string `json:"author_user"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content"`
}
