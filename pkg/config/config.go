package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

const EnvProd = "production"
const EnvLocal = "local"
const EnvConsumer = "consumer"

const prefix = "STEAM_"

var Config BaseConfig

func init() {

	// Set configs from environment variables
	Config.AdminUsername.Set("ADMIN_USER")
	Config.AdminPassword.Set("ADMIN_PASS")
	Config.AdminName.Set("ADMIN_NAME")
	Config.AdminEmail.Set("ADMIN_EMAIL")

	Config.RabbitUsername.Set("RABBIT_USER")
	Config.RabbitPassword.Set("RABBIT_PASS")
	Config.RabbitHost.Set("RABBIT_HOST")
	Config.RabbitPort.Set("RABBIT_PORT")
	Config.RabbitManagmentPort.Set("RABBIT_MANAGEMENT_PORT")

	Config.SessionAuthentication.Set("SESSION_AUTHENTICATION")
	Config.SessionEncryption.Set("SESSION_ENCRYPTION")

	Config.InstagramUsername.Set("INSTAGRAM_USERNAME")
	Config.InstagramPassword.Set("INSTAGRAM_PASSWORD")

	Config.MySQLHost.Set("MYSQL_HOST")
	Config.MySQLPort.Set("MYSQL_PORT")
	Config.MySQLUsername.Set("MYSQL_USERNAME")
	Config.MySQLPassword.Set("MYSQL_PASSWORD")
	Config.MySQLDatabase.Set("MYSQL_DATABASE")

	Config.RecaptchaPublic.Set("RECAPTCHA_PUBLIC")
	Config.RecaptchaPrivate.Set("RECAPTCHA_PRIVATE")

	Config.TwitchClientID.Set("TWITCH_CLIENT_ID")
	Config.TwitchClientSecret.Set("TWITCH_CLIENT_SECRET")

	Config.InfluxURL.Set("INFLUX_URL")
	Config.InfluxUsername.Set("INFLUX_USERNAME")
	Config.InfluxPassword.Set("INFLUX_PASSWORD")

	Config.MongoHost.Set("MONGO_HOST")
	Config.MongoPort.Set("MONGO_PORT")
	Config.MongoUsername.Set("MONGO_USERNAME")
	Config.MongoPassword.Set("MONGO_PASSWORD")
	Config.MongoDatabase.Set("MONGO_DATABASE")

	Config.TwitterAccessToken.Set("TWITTER_ACCESS_TOKEN")
	Config.TwitterAccessTokenSecret.Set("TWITTER_ACCESS_TOKEN_SECRET")
	Config.TwitterConsumerKey.Set("TWITTER_CONSUMER_KEY")
	Config.TwitterConsumerSecret.Set("TWITTER_CONSUMER_SECRET")

	Config.PatreonSecret.Set("PATREON_WEBOOK_SECRET")
	Config.PatreonClientID.Set("PATREON_CLIENT_ID")
	Config.PatreonClientSecret.Set("PATREON_CLIENT_SECRET")

	Config.DiscordClientID.Set("DISCORD_CLIENT_ID")
	Config.DiscordSescret.Set("DISCORD_SECRET")
	Config.DiscordBotToken.Set("DISCORD_BOT_TOKEN")
	Config.DiscordRelayToken.Set("DISCORD_RELAY_TOKEN")

	Config.GameDBDomain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.GithubToken.Set("GITHUB_TOKEN")
	Config.GoogleBucket.Set("GOOGLE_BUCKET")
	Config.GoogleProject.Set("GOOGLE_PROJECT")
	Config.GameDBDirectory.Set("PATH")
	Config.SendGridAPIKey.Set("SENDGRID")
	Config.SteamAPIKey.Set("API_KEY")
	Config.WebserverPort.Set("PORT")

	// Defaults
	Config.GameDBShortName.SetDefault("GameDB")
	Config.InstagramUsername.SetDefault("gamedb.online")
	Config.WebserverPort.SetDefault("8081")
	Config.EnableWebserver.SetDefault("1")
	Config.EnableConsumers.SetDefault("1")
	Config.GameDBDirectory.SetDefault("/root")
	Config.NewReleaseDays.SetDefault("14")

	switch Config.Environment.Get() {
	case EnvProd:

		Config.MemcacheDSN.SetDefault("memcache:11211")
		Config.EnableConsumers.SetDefault("0")

	case EnvLocal:

		Config.MemcacheDSN.SetDefault("localhost:11211")
		Config.PatreonSecret.SetDefault("EZTRjtID_1LUmgnQ4_WWuWIQbfj4QA1JtqYMq4prcq_kDvNdEXlgj2K7JyLwNXfd")

	case EnvConsumer:

		Config.EnableWebserver.SetDefault("0")

	default:
		fmt.Println("Missing env")
		os.Exit(1)
	}
}

type BaseConfig struct {
	AdminEmail    ConfigItem
	AdminName     ConfigItem
	AdminPassword ConfigItem
	AdminUsername ConfigItem

	DiscordClientID   ConfigItem
	DiscordSescret    ConfigItem
	DiscordRelayToken ConfigItem
	DiscordBotToken   ConfigItem

	InfluxURL      ConfigItem
	InfluxPassword ConfigItem
	InfluxUsername ConfigItem

	InstagramPassword ConfigItem
	InstagramUsername ConfigItem

	MongoHost     ConfigItem
	MongoPort     ConfigItem
	MongoUsername ConfigItem
	MongoPassword ConfigItem
	MongoDatabase ConfigItem

	MySQLHost     ConfigItem
	MySQLPort     ConfigItem
	MySQLUsername ConfigItem
	MySQLPassword ConfigItem
	MySQLDatabase ConfigItem

	RabbitUsername      ConfigItem
	RabbitPassword      ConfigItem
	RabbitHost          ConfigItem
	RabbitPort          ConfigItem
	RabbitManagmentPort ConfigItem

	RecaptchaPrivate ConfigItem
	RecaptchaPublic  ConfigItem

	SessionAuthentication ConfigItem
	SessionEncryption     ConfigItem

	TwitchClientID     ConfigItem
	TwitchClientSecret ConfigItem

	TwitterAccessToken       ConfigItem
	TwitterAccessTokenSecret ConfigItem
	TwitterConsumerKey       ConfigItem
	TwitterConsumerSecret    ConfigItem

	PatreonSecret       ConfigItem
	PatreonClientID     ConfigItem
	PatreonClientSecret ConfigItem

	Environment     ConfigItem
	GameDBDirectory ConfigItem
	GameDBDomain    ConfigItem
	GameDBShortName ConfigItem
	GithubToken     ConfigItem
	GoogleBucket    ConfigItem
	GoogleProject   ConfigItem
	MemcacheDSN     ConfigItem
	SendGridAPIKey  ConfigItem
	SteamAPIKey     ConfigItem
	WebserverPort   ConfigItem
	EnableWebserver ConfigItem
	EnableConsumers ConfigItem
	CommitHash      ConfigItem
	NewReleaseDays  ConfigItem
}

func (c BaseConfig) RabbitDSN() string {
	return "amqp://" + c.RabbitUsername.Get() + ":" + c.RabbitPassword.Get() + "@" + c.RabbitHost.Get() + ":" + c.RabbitPort.Get()
}

func (c BaseConfig) MySQLDNS() string {
	return c.MySQLUsername.Get() + ":" + c.MySQLPassword.Get() + "@tcp(" + c.MySQLHost.Get() + ":" + c.MySQLPort.Get() + ")/" + c.MySQLDatabase.Get()
}

func (c BaseConfig) MongoDSN() string {
	return "mongodb://" + c.MongoHost.Get() + ":" + c.MongoPort.Get()
}

func (c BaseConfig) RabbitAPI(values url.Values) string {
	return "http://" + c.RabbitHost.Get() + ":" + c.RabbitManagmentPort.Get() + "/api/overview?" + values.Encode()
}

func (c BaseConfig) ListenOn() string {
	return "0.0.0.0:" + c.WebserverPort.Get()
}

func (c BaseConfig) HasMemcache() bool {
	return c.MemcacheDSN.Get() != ""
}

func (c BaseConfig) IsLocal() bool {
	return c.Environment.Get() == EnvLocal
}

func (c BaseConfig) IsProd() bool {
	return c.Environment.Get() == EnvProd
}

func (c BaseConfig) IsConsumer() bool {
	return c.Environment.Get() == EnvConsumer
}

// ConfigItem
type ConfigItem struct {
	value        string
	defaultValue string
}

func (ci *ConfigItem) Set(environment string) {
	env, b := os.LookupEnv(prefix + environment)
	if !b {
		fmt.Println("MISSING ENV: " + environment)
	}
	ci.value = env
}

func (ci *ConfigItem) SetDefault(defaultValue string) {
	ci.defaultValue = defaultValue
}

func (ci ConfigItem) Get() string {
	if ci.value != "" {
		return ci.value
	}
	return ci.defaultValue
}

func (ci ConfigItem) GetBool() bool {
	b, err := strconv.ParseBool(ci.Get())
	if err != nil {
		fmt.Println(err)
	}
	return b
}

func (ci ConfigItem) GetInt() int {
	i, err := strconv.Atoi(ci.Get())
	if err != nil {
		fmt.Println(err)
	}
	return i
}
