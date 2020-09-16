package pages

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"github.com/nlopes/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func WebhooksRouter() http.Handler {

	r := chi.NewRouter()
	r.Post("/patreon", patreonWebhookPostHandler)
	r.Post("/github", gitHubWebhookPostHandler)
	r.Post("/twitter", twitterWebhookPostHandler)
	r.Post("/sendgrid", sendgridWebhookPostHandler)
	return r
}

func sendgridWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	var valid bool

	if r.Header.Get("X-Twilio-Email-Event-Webhook-Signature") == config.C.SendGridSecret {
		valid = true
	}

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), 500)
		return
	}

	defer helpers.Close(r.Body)

	zap.L().Named(log.LogNameSendGrid).Debug("SendGrid webhook", zap.ByteString("body", body), zap.Bool("valid", valid))
}

func twitterWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	secret := r.Header.Get("secret")

	zap.L().Named(log.LogNameTwitter).Debug("Twitter webhook", zap.String("secret", secret))

	if config.C.TwitterZapierSecret == "" {

		log.Fatal("Missing environment variables")
		return
	}

	if config.C.TwitterZapierSecret != secret {
		return
	}

	// Get body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), 500)
		return
	}

	defer helpers.Close(r.Body)

	webhooks := twitterWebhook{}
	err = json.Unmarshal(b, &webhooks)

	//
	if webhooks.Name == "gamedb_online" && webhooks.OriginalName == "" {

		// Delete cache
		err = memcache.Delete(memcache.HomeTweets.Key)
		if err != nil {
			log.Err(err.Error())
		}

		// Forward to Discord
		discordSession, err := discordgo.New("Bot " + config.C.DiscordRelayBotToken)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		_, err = discordSession.ChannelMessageSend("407493777058693121", webhooks.URL)
		if err != nil {
			log.Err(err.Error())
		}
	}

	returnJSON(w, r, nil)
}

type twitterWebhook struct {
	Name         string `json:"screen_name"`
	OriginalName string `json:"retweeted_screen_name"`
	Text         string `json:"full_text"`
	URL          string `json:"url"`
}

const (
	PATREON_TIER_1 = 2431311
	PATREON_TIER_2 = 2431320
	PATREON_TIER_3 = 2431347
)

func patreonWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	b, event, err := patreon.Validate(r, config.C.PatreonSecret)
	if err != nil {
		log.ErrS(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = mongo.InsertOne(mongo.CollectionPatreonWebhooks, mongo.PatreonWebhook{
		CreatedAt:   time.Now(),
		RequestBody: string(b),
		Event:       event,
	})
	if err != nil {
		log.ErrS(err)
	}

	pwr, err := patreon.Unmarshal(b)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("webhook", b))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = savePatreonWebhookEvent(r, mongo.EventEnum(event), pwr)
	if err != nil {
		log.ErrS(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Slack message
	if config.C.SlackPatreonWebhook == "" {
		log.Fatal("Missing environment variables")
	} else {
		err = slack.PostWebhook(config.C.SlackPatreonWebhook, &slack.WebhookMessage{Text: event})
		log.ErrS(err)
	}

	returnJSON(w, r, nil)
}

func savePatreonWebhookEvent(r *http.Request, event mongo.EventEnum, pwr patreon.Webhook) (err error) {

	email := pwr.User.Attributes.Email
	if email == "" {
		return nil
	}

	player := mongo.Player{}
	err = mongo.FindOne(mongo.CollectionPlayers, bson.D{{Key: "email", Value: email}}, nil, bson.M{"_id": 1}, &player)
	if err == mongo.ErrNoDocuments {
		return nil
	}
	if err != nil {
		return err
	}

	user, err := mysql.GetUserByKey("steam_id", player.ID, 0)
	if err == mysql.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	return mongo.CreateUserEvent(r, user.ID, mongo.EventPatreonWebhook+"-"+event)
}

const signaturePrefix = "sha1="
const signatureLength = 45

func gitHubWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), 500)
		return
	}

	defer helpers.Close(r.Body)

	//
	var signature = r.Header.Get("X-Hub-Signature")
	var event = r.Header.Get("X-GitHub-Event")

	zap.L().Named(log.LognameGitHub).Debug("Incoming GitHub webhook", zap.ByteString("webhook", body), zap.String("event", event))

	if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
		http.Error(w, "Invalid signature (1)", 400)
		return
	}

	mac := hmac.New(sha1.New, []byte(config.C.GithubWebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signaturePrefix+expectedMAC), []byte(signature)) {
		log.Err("Invalid signature (2)", zap.String("secret", config.C.GithubWebhookSecret))
		http.Error(w, "Invalid signature (2)", 400)
		return
	}

	switch event {
	case "push":

		// Clear cache
		items := []string{
			memcache.MemcacheCommitsPage(1).Key,
			memcache.MemcacheCommitsTotal.Key,
		}

		err := memcache.Delete(items...)
		if err != nil {
			log.ErrS(err)
		}
	}

	returnJSON(w, r, nil)
}
