package pages

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

const (
	chatBotClientID = "567257603185311745"
)

var discordChatBotSession *discordgo.Session

func init() {

	var err error

	discordChatBotSession, err = discordgo.New("Bot " + config.Config.DiscordChatBotToken.Get())
	if err != nil {
		log.Err(err)
		panic(err)
	}
}

func ChatBotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", chatBotHandler)
	return r
}

func chatBotHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := chatBotTemplate{}
	t.fill(w, r, "Chat", "The Game DB community.")
	t.Commands = t.CommandsGrouped()

	// Get amount of guilds, todo, cache
	func() {

		after := ""
		more := true

		for more {

			guilds, err := discordChatBotSession.UserGuilds(100, "", after)
			if err != nil {
				log.Warning(err)
				return
			}

			if len(guilds) < 100 {
				more = false
			}

			for _, v := range guilds {
				after = v.ID
			}

			t.Guilds = t.Guilds + len(guilds)
		}

		log.Info(strconv.Itoa(t.Guilds) + " guilds")
	}()

	returnTemplate(w, r, "chat_bot", t)
}

type chatBotTemplate struct {
	GlobalTemplate
	Commands [][]chatbot.Command
	Guilds   int
}

func (cbt chatBotTemplate) AddBotLink() string {
	return "https://discordapp.com/oauth2/authorize?client_id=" + chatBotClientID + "&scope=bot&permissions=0"
}

func (cbt chatBotTemplate) CommandsGrouped() (ret [][]chatbot.Command) {

	var groupedMap = map[chatbot.CommandType][]chatbot.Command{}
	for _, v := range chatbot.CommandRegister {

		if _, ok := groupedMap[v.Type()]; ok {

			groupedMap[v.Type()] = append(groupedMap[v.Type()], v)

		} else {

			groupedMap[v.Type()] = []chatbot.Command{v}

		}
	}

	for _, v := range groupedMap {

		sort.Slice(v, func(i, j int) bool {
			return v[i].Example() < v[j].Example()
		})

		ret = append(ret, v)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i][0].Type() < ret[j][0].Type()
	})

	return ret
}
