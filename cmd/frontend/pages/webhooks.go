package pages

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"golang.org/x/text/currency"
)

func WebhooksRouter() http.Handler {

	r := chi.NewRouter()
	r.Post("/patreon", patreonWebhookPostHandler)
	r.Post("/github", gitHubWebhookPostHandler)
	r.Post("/twitter", twitterZapierWebhookPostHandler)
	r.Post("/sendgrid", sendgridWebhookPostHandler)
	r.Post("/mailjet", mailjetWebhookPostHandler)
	return r
}

func mailjetWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Save webhook
	err = mongo.NewWebhook(mongo.WebhookServiceMailjet, "", string(body))
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func sendgridWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Validate
	if config.C.SendGridSecret == "" {
		log.ErrS("Missing sendgrid environment variables")
	}

	if r.Header.Get("X-Twilio-Email-Event-Webhook-Signature") != config.C.SendGridSecret {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Save webhook
	err = mongo.NewWebhook(mongo.WebhookServiceSendgrid, "", string(body))
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func twitterZapierWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Validate
	if config.C.TwitterZapierSecret == "" {
		log.ErrS("Missing zapier environment variables")
	}

	if config.C.TwitterZapierSecret != r.Header.Get("secret") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Save webhook
	err = mongo.NewWebhook(mongo.WebhookServiceTwitter, "", string(body))
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle
	webhooks := twitterWebhook{}
	err = json.Unmarshal(body, &webhooks)

	if webhooks.Name == "gamedb_online" && webhooks.OriginalName == "" {

		// Delete cache
		err = memcache.Delete(memcache.HomeTweets.Key)
		if err != nil {
			log.Err(err.Error())
		}

		// Forward to Discord
		if config.C.DiscordRelayBotToken == "" {
			log.ErrS("Missing discord environment variable")
		}

		discordSession, err := discordgo.New("Bot " + config.C.DiscordRelayBotToken)
		if err != nil {
			log.ErrS(err)
			return
		}

		_, err = discordSession.ChannelMessageSend(generalChannelID, webhooks.URL)
		if err != nil {
			log.Err(err.Error())
		}
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
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

	// Validate
	if config.C.PatreonSecret == "" {
		log.ErrS("Missing patreon environment variable")
	}

	b, event, err := patreon.Validate(r, config.C.PatreonSecret)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save webhook
	err = mongo.NewWebhook(mongo.WebhookServicePatreon, event, string(b))
	if err != nil {
		log.ErrS(err)
	}

	// Handle
	pwr, err := patreon.Unmarshal(b)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("webhook", b))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find user
	user, err := mysql.GetUserByProviderID(oauth.ProviderPatreon, strconv.Itoa(int(pwr.User.ID)))
	if err == mysql.ErrRecordNotFound {

		user, err = mysql.GetUserByEmail(pwr.User.Attributes.Email)
		if err != nil {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update donation bits
	if user.Donated < pwr.Data.Attributes.LifetimeSupportCents {

		// Get player ID
		var playerID int64

		steam, err := mysql.GetUserProviderByUserID(oauth.ProviderSteam, user.ID)
		if err != nil && err != mysql.ErrRecordNotFound {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if err == nil {
			playerID, err = strconv.ParseInt(steam.ID, 10, 64)
			if err != nil {
				log.ErrS(err)
			}
		}

		// Save donation
		db, err := mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		donation := mysql.Donation{
			UserID:           user.ID,
			PlayerID:         playerID,
			Email:            user.Email,
			AmountUSD:        pwr.Data.Attributes.WillPayAmountCents,
			OriginalCurrency: currency.USD.String(),
			OriginalAmount:   pwr.Data.Attributes.WillPayAmountCents,
			Source:           mysql.DonationSourcePatreon,
			Anon:             false, // todo
		}

		db = db.Create(&donation)
		if db.Error != nil {
			log.ErrS(db.Error)
			http.Error(w, db.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Update user
		db, err = mysql.GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		update := map[string]interface{}{
			"donated": pwr.Data.Attributes.LifetimeSupportCents,
		}

		db = db.Model(&user).Updates(update)
		if db.Error != nil {
			log.ErrS(db.Error)
			http.Error(w, db.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Create event
		err = mongo.NewEvent(r, user.ID, mongo.EventPatreonWebhook+"-"+mongo.EventEnum(event))
		if err != nil {
			log.Err(err.Error(), zap.ByteString("webhook", b))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}

func gitHubWebhookPostHandler(w http.ResponseWriter, r *http.Request) {

	// Get body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer helpers.Close(r.Body)

	// Validate
	var signature = r.Header.Get("X-Hub-Signature")
	var event = r.Header.Get("X-GitHub-Event")

	if len(signature) != 45 || !strings.HasPrefix(signature, "sha1=") {
		http.Error(w, "Invalid signature (1)", 400)
		return
	}

	if config.C.GithubWebhookSecret == "" {
		log.ErrS("Missing github environment variables")
	}

	mac := hmac.New(sha1.New, []byte(config.C.GithubWebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte("sha1="+expectedMAC), []byte(signature)) {
		log.Err("Invalid signature (2)", zap.String("secret", config.C.GithubWebhookSecret))
		http.Error(w, "Invalid signature (2)", 400)
		return
	}

	// Save webhook
	err = mongo.NewWebhook(mongo.WebhookServiceGithub, event, string(body))
	if err != nil {
		log.ErrS(err)
	}

	// Handle
	switch event {
	case "push":

		// Clear cache
		err := memcache.Delete(memcache.MemcacheCommitsPage(1).Key)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Return
	_, err = w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.ErrS(err)
	}
}
