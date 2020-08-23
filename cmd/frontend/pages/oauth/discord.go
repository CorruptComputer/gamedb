package oauth

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
	"golang.org/x/oauth2"
)

type discordConnection struct {
	baseConnection
}

func (c discordConnection) getID(r *http.Request, token *oauth2.Token) (string, error) {

	discord, err := discordgo.New("Bearer " + token.AccessToken)
	if err != nil {
		return "", err
	}

	discordUser, err := discord.User("@me")
	if err != nil {
		return "", oauthError{err, "An error occurred (1003)"}
	}

	// if !discordUser.Verified { // Seems ot always be false
	// 	err = session.SetFlash(r, helpers.SessionBad, "This Discord account has not been verified")
	// 	zap.S().Error(err)
	// 	return
	// }

	return discordUser.ID, nil
}

func (c discordConnection) getName() string {
	return "Discord"
}

func (c discordConnection) getEnum() ConnectionEnum {
	return ConnectionDiscord
}

func (c discordConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.C.GameDBDomain + "/login/oauth-callback/discord"
	} else {
		redirectURL = config.C.GameDBDomain + "/settings/oauth-callback/discord"
	}

	return oauth2.Config{
		ClientID:     config.C.DiscordClientID,
		ClientSecret: config.C.DiscordClientSescret,
		Scopes:       []string{"identify"},
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
}

func (c discordConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {
	c.linkOAuth(w, r, c, false)
}

func (c discordConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	c.unlink(w, r, c, mongo.EventUnlinkDiscord)
}

func (c discordConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLinkDiscord, false)

	session.Save(w, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (c discordConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {
	c.linkOAuth(w, r, c, true)
}

func (c discordConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
