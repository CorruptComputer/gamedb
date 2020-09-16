package mongo

import (
	"time"

	"github.com/Jleagle/patreon-go/patreon"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type WebhookService string

const (
	WebhookServicePatreon  WebhookService = "patreon"
	WebhookServiceGithub   WebhookService = "github"
	WebhookServiceTwitter  WebhookService = "twitter"
	WebhookServiceSendgrid WebhookService = "sendgrid"
	WebhookServiceMailjet  WebhookService = "mailjet"
)

type Webhook struct {
	CreatedAt   time.Time      `bson:"created_at"`
	RequestBody string         `bson:"request_body"`
	Event       string         `bson:"event"`
	Service     WebhookService `bson:"service"`
}

func (webhook Webhook) BSON() bson.D {

	return bson.D{
		{"created_at", webhook.CreatedAt},
		{"request_body", webhook.RequestBody},
		{"event", webhook.Event},
		{"service", webhook.Service},
	}
}

func (webhook Webhook) UnmarshalPatreon() (wh patreon.Webhook, err error) {

	err = helpers.Unmarshal([]byte(webhook.RequestBody), &wh)
	return wh, err
}

func NewWebhook(service WebhookService, event string, body string) error {

	row := Webhook{
		CreatedAt:   time.Now(),
		RequestBody: body,
		Event:       event,
		Service:     WebhookServicePatreon,
	}

	_, err := InsertOne(CollectionWebhooks, row)
	return err
}

func GetWebhooks(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (webhooks []Webhook, err error) {

	cur, ctx, err := Find(CollectionWebhooks, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return webhooks, err
	}

	defer close(cur, ctx)

	for cur.Next(ctx) {

		var webhook Webhook
		err := cur.Decode(&webhook)
		if err != nil {
			log.ErrS(err)
		} else {
			webhooks = append(webhooks, webhook)
		}
	}

	return webhooks, cur.Err()
}
