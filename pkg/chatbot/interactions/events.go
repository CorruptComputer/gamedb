package interactions

import (
	"strings"
	"time"
)

type Event struct {
	ChannelID string `json:"channel_id"`
	Data      struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Options []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"options"`
	} `json:"data"`
	GuildID string `json:"guild_id"`
	ID      string `json:"id"`
	Member  struct {
		Deaf         bool        `json:"deaf"`
		IsPending    bool        `json:"is_pending"`
		JoinedAt     time.Time   `json:"joined_at"`
		Mute         bool        `json:"mute"`
		Nick         interface{} `json:"nick"`
		Pending      bool        `json:"pending"`
		Permissions  string      `json:"permissions"`
		PremiumSince interface{} `json:"premium_since"`
		Roles        []string    `json:"roles"`
		User         struct {
			Avatar        string `json:"avatar"`
			Discriminator string `json:"discriminator"`
			ID            string `json:"id"`
			PublicFlags   int    `json:"public_flags"`
			Username      string `json:"username"`
		} `json:"user"`
	} `json:"member"`
	Token   string `json:"token"`
	Type    int    `json:"type"`
	Version int    `json:"version"`
}

func (e Event) Arguments() (a map[string]string) {

	a = map[string]string{}
	for _, v := range e.Data.Options {
		a[v.Name] = v.Value
	}
	return a
}

func (e Event) ArgumentsString() string {

	var s = []string{e.Data.Name}
	for _, v := range e.Data.Options {
		s = append(s, v.Name+":"+v.Value)
	}
	return strings.Join(s, " ")
}
