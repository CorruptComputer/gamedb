package oauth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	gh "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubConnection struct {
	baseConnection
}

func (c githubConnection) getID(r *http.Request, token *oauth2.Token) (string, error) {

	ctx := context.Background()

	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token.AccessToken,
		},
	)))

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(user.GetID(), 10), nil
}

func (c githubConnection) getName() string {
	return "GitHub"
}

func (c githubConnection) getEnum() ConnectionEnum {
	return ConnectionGithub
}

func (c githubConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.Config.GameDBDomain.Get() + "/login/oauth-callback/github"
	} else {
		redirectURL = config.Config.GameDBDomain.Get() + "/settings/oauth-callback/github"
	}

	return oauth2.Config{
		ClientID:     config.Config.GitHubClient.Get(),
		ClientSecret: config.Config.GitHubSecret.Get(),
		Scopes:       []string{""},
		RedirectURL:  redirectURL,
		Endpoint:     github.Endpoint,
	}
}

func (c githubConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {

	c.linkOAuth(w, r, c, false)
}

func (c githubConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {

	c.unlink(w, r, c, mongo.EventUnlinkGitHub)
}

func (c githubConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLinkGitHub, false)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (c githubConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	c.linkOAuth(w, r, c, true)
}

func (c githubConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
