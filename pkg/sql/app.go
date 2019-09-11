package sql

import (
	"errors"
	"html/template"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/jinzhu/gorm"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"
)

var (
	ErrInvalidAppID = errors.New("invalid id")
)

type App struct {
	Achievements                  string    `gorm:"not null;column:achievements;type:text"`           // []AppAchievement
	Background                    string    `gorm:"not null;column:background"`                       //
	BundleIDs                     string    `gorm:"not null;column:bundle_ids"`                       // []int
	Categories                    string    `gorm:"not null;column:categories;type:json"`             // []int8
	ChangeNumber                  int       `gorm:"not null;column:change_number"`                    //
	ChangeNumberDate              time.Time `gorm:"not null;column:change_number_date;type:datetime"` //
	ClientIcon                    string    `gorm:"not null;column:client_icon"`                      //
	ComingSoon                    bool      `gorm:"not null;column:coming_soon"`                      //
	Common                        string    `gorm:"not null;column:common"`                           // PICSAppCommon
	Config                        string    `gorm:"not null;column:config"`                           // PICSAppConfig
	CreatedAt                     time.Time `gorm:"not null;column:created_at;type:datetime"`         //
	Depots                        string    `gorm:"not null;column:depots"`                           // Depots
	Developers                    string    `gorm:"not null;column:developers;type:json"`             // []int
	DemoIDs                       string    `gorm:"not null;column:demo_ids;type:json"`               // []int
	DLC                           string    `gorm:"not null;column:dlc;type:json"`                    // []int
	DLCCount                      int       `gorm:"not null;column:dlc_count"`                        //
	Extended                      string    `gorm:"not null;column:extended"`                         // PICSExtended
	GameID                        int       `gorm:"not null;column:game_id"`                          //
	GameName                      string    `gorm:"not null;column:game_name"`                        //
	Genres                        string    `gorm:"not null;column:genres;type:json"`                 // []int
	GroupID                       string    `gorm:"not null;column:group_id;type:varchar"`            //
	Homepage                      string    `gorm:"not null;column:homepage"`                         //
	Icon                          string    `gorm:"not null;column:icon"`                             //
	ID                            int       `gorm:"not null;column:id;primary_key"`                   //
	Install                       string    `gorm:"not null;column:install"`                          // map[string]interface{}
	IsFree                        bool      `gorm:"not null;column:is_free;type:tinyint(1)"`          //
	Items                         int       `gorm:"not null;column:items;type:int"`                   //
	ItemsDigest                   string    `gorm:"not null;column:items_digest"`                     //
	Launch                        string    `gorm:"not null;column:launch"`                           // []db.PICSAppConfigLaunchItem
	Localization                  string    `gorm:"not null;column:localization"`                     // map[string]interface{}
	Logo                          string    `gorm:"not null;column:logo"`                             //
	MetacriticScore               int8      `gorm:"not null;column:metacritic_score"`                 //
	MetacriticURL                 string    `gorm:"not null;column:metacritic_url"`                   //
	Movies                        string    `gorm:"not null;column:movies;type:text"`                 // []AppVideo
	Name                          string    `gorm:"not null;column:name"`                             //
	NewsIDs                       string    `gorm:"not null;column:news_ids"`                         // []int64
	Packages                      string    `gorm:"not null;column:packages;type:json"`               // []int
	Platforms                     string    `gorm:"not null;column:platforms;type:json"`              // []string
	PlayerPeakWeek                int       `gorm:"not null;column:player_peak_week"`                 //
	PlayerPeakAllTime             int       `gorm:"not null;column:player_peak_alltime"`              //
	PlayerAverageWeek             float64   `gorm:"not null;column:player_avg_week;type:float"`       //
	PlayerTrend                   int64     `gorm:"not null;column:player_trend"`                     //
	Prices                        string    `gorm:"not null;column:prices"`                           // ProductPrices
	PublicOnly                    bool      `gorm:"not null;column:public_only"`                      //
	Publishers                    string    `gorm:"not null;column:publishers;type:json"`             // []int
	RelatedAppIDs                 string    `gorm:"not null;column:related_app_ids;type:json"`        // []int
	ReleaseDate                   string    `gorm:"not null;column:release_date"`                     //
	ReleaseDateUnix               int64     `gorm:"not null;column:release_date_unix"`                //
	ReleaseState                  string    `gorm:"not null;column:release_state"`                    //
	Reviews                       string    `gorm:"not null;column:reviews"`                          // AppReviewSummary
	ReviewsScore                  float64   `gorm:"not null;column:reviews_score"`                    //
	Screenshots                   string    `gorm:"not null;column:screenshots;type:text"`            // []AppImage
	ShortDescription              string    `gorm:"not null;column:description_short"`                //
	Stats                         string    `gorm:"not null;column:stats;type:text"`                  // []AppStat
	SteamSpy                      string    `gorm:"not null;column:steam_spy"`                        // AppSteamSpy
	SystemRequirements            string    `gorm:"not null;column:system_requirements"`              // map[string]interface{}
	Tags                          string    `gorm:"not null;column:tags;type:json"`                   // []int
	Type                          string    `gorm:"not null;column:type"`                             //
	TwitchID                      int       `gorm:"not null;column:twitch_id"`                        //
	UFS                           string    `gorm:"not null;column:ufs"`                              // PICSAppUFS
	UpdatedAt                     time.Time `gorm:"not null;column:updated_at;type:datetime"`         //
	Version                       string    `gorm:"not null;column:version"`                          //
	AchievementsCount             int       `gorm:"not null;column:achievements_count"`               //
	AchievementsAverageCompletion float64   `gorm:"not null;column:achievements_average_completion"`  //
	PlaytimeTotal                 int64     `gorm:"not null;column:playtime_total"`                   // Minutes
	PlaytimeAverage               float64   `gorm:"not null;column:playtime_average"`                 // Minutes
}

func (app *App) BeforeCreate(scope *gorm.Scope) error {
	return app.Before(scope)
}

func (app *App) BeforeSave(scope *gorm.Scope) error {
	return app.Before(scope)
}

func (app *App) Before(scope *gorm.Scope) error {

	if app.Achievements == "" {
		app.Achievements = "[]"
	}
	if app.BundleIDs == "" || app.BundleIDs == "null" {
		app.BundleIDs = "[]"
	}
	if app.Categories == "" {
		app.Categories = "[]"
	}
	if app.ChangeNumberDate.IsZero() {
		app.ChangeNumberDate = time.Now()
	}
	if app.Common == "" {
		app.Common = "{}"
	}
	if app.Config == "" {
		app.Config = "{}"
	}
	if app.Depots == "" {
		app.Depots = "{}"
	}
	if app.Developers == "" {
		app.Developers = "[]"
	}
	if app.DemoIDs == "" {
		app.DemoIDs = "[]"
	}
	if app.DLC == "" {
		app.DLC = "[]"
	}
	if app.Extended == "" {
		app.Extended = "{}"
	}
	if app.Genres == "" {
		app.Genres = "[]"
	}
	if app.Install == "" {
		app.Install = "{}"
	}
	if app.Launch == "" {
		app.Launch = "[]"
	}
	if app.Localization == "" {
		app.Localization = "{}"
	}
	if app.Movies == "" {
		app.Movies = "[]"
	}
	if app.NewsIDs == "" {
		app.NewsIDs = "[]"
	}
	if app.Packages == "" {
		app.Packages = "[]"
	}
	if app.Platforms == "" {
		app.Platforms = "[]"
	}
	if app.Prices == "" {
		app.Prices = "{}"
	}
	if app.Publishers == "" {
		app.Publishers = "[]"
	}
	if app.Reviews == "" {
		app.Reviews = "{}"
	}
	if app.Stats == "" {
		app.Stats = "[]"
	}
	if app.Screenshots == "" {
		app.Screenshots = "[]"
	}
	if app.SteamSpy == "" {
		app.SteamSpy = "{}"
	}
	if app.SystemRequirements == "" {
		app.SystemRequirements = "{}"
	}
	if app.Tags == "" {
		app.Tags = "[]"
	}
	if app.UFS == "" {
		app.UFS = "{}"
	}

	return nil
}

func (app App) GetID() int {
	return app.ID
}

func (app App) GetProductType() helpers.ProductType {
	return helpers.ProductTypeApp
}

func (app App) GetPath() string {
	return helpers.GetAppPath(app.ID, app.Name)
}

func (app App) GetType() (ret string) {

	switch app.Type {
	case "dlc":
		return "DLC"
	case "":
		return "Unknown"
	default:
		return strings.Title(app.Type)
	}
}

func (app App) GetReviewScore() string {

	return helpers.FloatToString(app.ReviewsScore, 2) + "%"
}

func (app App) GetDaysToRelease() string {

	return helpers.GetDaysToRelease(app.ReleaseDateUnix)
}

func (app App) GetReleaseState() (ret string) {

	switch app.ReleaseState {
	case "preloadonly":
		return "Preload Only"
	case "prerelease":
		return "Pre Release"
	case "released":
		return "Released"
	case "":
		return "Unreleased"
	default:
		return strings.Title(app.ReleaseState)
	}
}

func (app App) GetReleaseDateNice() string {

	if app.ReleaseDateUnix == 0 {
		return app.ReleaseDate
	}

	return time.Unix(app.ReleaseDateUnix, 0).Format(helpers.DateYear)
}

func (app App) GetUpdatedNice() string {
	return app.UpdatedAt.Format(helpers.DateYearTime)
}

func (app App) GetPICSUpdatedNice() string {

	d := app.ChangeNumberDate

	// 0000-01-01 00:00:00
	if d.Unix() == -62167219200 {
		return "-"
	}

	if d.IsZero() {
		return "-"
	}
	return d.Format(helpers.DateYearTime)
}

func (app App) GetIcon() (ret string) {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPrices() (prices ProductPrices, err error) {

	err = helpers.Unmarshal([]byte(app.Prices), &prices)

	// Needed for marshalling into array
	if len(prices) == 0 {
		prices = ProductPrices{}
	}

	return prices, err
}

func (app App) GetPrice(code steam.ProductCC) (price ProductPrice) {

	prices, err := app.GetPrices()
	if err != nil {
		return price
	}

	return prices.Get(code)
}

func (app App) GetNewsIDs() (ids []int64, err error) {

	if app.NewsIDs == "" {
		return ids, err
	}

	err = helpers.Unmarshal([]byte(app.NewsIDs), &ids)
	return ids, err
}

func (app App) GetExtended() (extended pics.PICSKeyValues) {

	extended = pics.PICSKeyValues{}

	err := helpers.Unmarshal([]byte(app.Extended), &extended)
	log.Err(err)

	return extended
}

func (app App) GetCommon() (common pics.PICSKeyValues) {

	common = pics.PICSKeyValues{}
	err := helpers.Unmarshal([]byte(app.Common), &common)
	log.Err(err)

	return common
}

func (app App) GetConfig() (config pics.PICSKeyValues) {

	config = pics.PICSKeyValues{}
	err := helpers.Unmarshal([]byte(app.Config), &config)
	log.Err(err)

	return config
}

func (app App) GetUFS() (ufs pics.PICSKeyValues) {

	ufs = pics.PICSKeyValues{}
	err := helpers.Unmarshal([]byte(app.UFS), &ufs)
	log.Err(err)

	return ufs
}

func (app App) GetDepots() (depots pics.Depots, err error) {

	err = helpers.Unmarshal([]byte(app.Depots), &depots)
	log.Err(err)
	return depots, err
}

func (app App) GetLaunch() (items []pics.PICSAppConfigLaunchItem, err error) {

	err = helpers.Unmarshal([]byte(app.Launch), &items)
	log.Err(err)
	return items, err
}

func (app App) GetInstall() (install map[string]interface{}, err error) {

	install = map[string]interface{}{}

	err = helpers.Unmarshal([]byte(app.Install), &install)
	log.Err(err)
	return install, err
}

func (app App) GetLocalization() (localization pics.Localisation) {

	localization = pics.Localisation{}
	err := helpers.Unmarshal([]byte(app.Localization), &localization)
	log.Err(err)

	return localization
}

func (app App) GetSystemRequirements() (ret []SystemRequirement, err error) {

	systemRequirements := map[string]interface{}{}

	err = helpers.Unmarshal([]byte(app.SystemRequirements), &systemRequirements)
	log.Err(err)

	flattened := helpers.FlattenMap(systemRequirements)

	for k, v := range flattened {
		if val, ok := v.(string); ok {
			ret = append(ret, SystemRequirement{Key: k, Val: val})
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})

	return ret, err
}

type SystemRequirement struct {
	Key string
	Val string
}

func (sr SystemRequirement) Format() template.HTML {

	switch sr.Val {
	case "0":
		return `<i class="fas fa-times text-danger"></i>`
	case "1":
		return `<i class="fas fa-check text-success"></i>`
	case "warn":
		return `<span class="text-warning">Warn</span>`
	case "deny":
		return `<span class="text-danger">Deny</span>`
	default:
		return template.HTML(sr.Val)
	}
}

func (app App) IsOnSale() bool {

	common := app.GetCommon()

	if common.GetValue("app_retired_publisher_request") == "1" {
		return false
	}

	return true
}

func (app App) GetOnlinePlayers() (players int64, err error) {

	var item = helpers.MemcacheAppPlayersRow(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("player_count", "")
		builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", app.ID)
		builder.AddOrderBy("time", false)
		builder.SetLimit(1)

		resp, err := helpers.InfluxQuery(builder.String())

		return helpers.GetFirstInfluxInt(resp), err
	})

	return players, err
}

func (app App) GetCommunityLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func (app App) GetStoreLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://store.steampowered.com/app/" + strconv.Itoa(app.ID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func (app App) GetPCGamingWikiLink() string {
	return "https://pcgamingwiki.com/api/appid.php?appid=" + strconv.Itoa(app.ID)
}

func (app App) GetHeaderImage() string {
	return "https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(app.ID) + "/header.jpg"
}

func (app App) GetHeaderImage2() string {

	params := url.Values{}
	params.Set("url", app.GetHeaderImage())
	params.Set("q", "10")

	return "https://images.weserv.nl?" + params.Encode()
}

func (app App) GetInstallLink() template.URL {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetMetacriticLink() template.URL {
	return template.URL("https://www.metacritic.com/game/" + app.MetacriticURL)
}

func (app App) GetScreenshots() (screenshots []AppImage, err error) {

	err = helpers.Unmarshal([]byte(app.Screenshots), &screenshots)
	log.Err(err)
	return screenshots, err
}

func (app App) GetMovies() (movies []AppVideo, err error) {

	err = helpers.Unmarshal([]byte(app.Movies), &movies)
	log.Err(err)
	return movies, err
}

func (app App) GetSteamSpy() (ss AppSteamSpy, err error) {

	err = helpers.Unmarshal([]byte(app.SteamSpy), &ss)
	log.Err(err)
	return ss, err
}

func (app App) GetCoopTags() (string, error) {

	tags, err := app.GetTagIDs()
	if err != nil {
		return "", err
	}

	var tagMap = map[int]string{
		1685: "Co-op",
		3843: "Online co-op",
		3841: "Local co-op",
		4508: "Co-op campaign",

		3859:  "Multiplayer",
		128:   "Massively multiplayer",
		7368:  "Local multiplayer",
		17770: "Asynchronous multiplayer",
	}

	var coopTags []string
	for _, v := range tags {
		if val, ok := tagMap[v]; ok {
			coopTags = append(coopTags, val)
		}
	}

	return strings.Join(coopTags, ", "), nil
}

// Template
func (app App) GetAchievements() (achievements []AppAchievement, err error) {

	err = helpers.Unmarshal([]byte(app.Achievements), &achievements)
	return achievements, err
}

func (app App) GetStats() (stats []AppStat, err error) {

	err = helpers.Unmarshal([]byte(app.Stats), &stats)
	return stats, err
}

func (app App) GetDemoIDs() (demos []int, err error) {

	err = helpers.Unmarshal([]byte(app.DemoIDs), &demos)
	return demos, err
}

func (app App) GetDemos() (demos []App, err error) {

	var item = helpers.MemcacheAppDemos(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &demos, func() (interface{}, error) {

		ids, err := app.GetDemoIDs()
		if err != nil {
			return demos, err
		}

		return GetAppsByID(ids, []string{"id", "name"})
	})

	if len(demos) == 0 {
		demos = []App{} // Needed for marshalling into type
	}

	return demos, err
}

func (app App) GetPlatforms() (platforms []string, err error) {

	err = helpers.Unmarshal([]byte(app.Platforms), &platforms)
	return platforms, err
}

func (app App) GetPlatformImages() (ret template.HTML, err error) {

	if app.Platforms == "" {
		return template.HTML(""), nil
	}

	platforms, err := app.GetPlatforms()
	if err != nil {
		log.Err(err)
		return ret, err
	}

	if helpers.SliceHasString(platforms, platformWindows) {
		ret = ret + `<a href="/apps?platforms=windows"><i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformMac) {
		ret = ret + `<a href="/apps?platforms=macos"><i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformLinux) {
		ret = ret + `<a href="/apps?platforms=linux"><i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret, nil
}

func (app App) GetDLCIDs() (dlcs []int, err error) {

	err = helpers.Unmarshal([]byte(app.DLC), &dlcs)
	log.Err(err)

	if len(dlcs) == 0 {
		dlcs = []int{} // Needed for marshalling into type
	}

	return dlcs, err
}

func (app App) GetDLCs() (apps []App, err error) {

	var item = helpers.MemcacheAppDLC(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		ids, err := app.GetDLCIDs()
		if err != nil {
			return apps, err
		}

		return GetAppsByID(ids, []string{"id", "name"})
	})

	if len(apps) == 0 {
		apps = []App{} // Needed for marshalling into type
	}

	return apps, err
}

func (app App) GetPackageIDs() (packages []int, err error) {

	packages = []int{} // Needed for marshalling into type

	err = helpers.Unmarshal([]byte(app.Packages), &packages)
	log.Err(err)
	return packages, err
}

func (app App) GetReviews() (reviews AppReviewSummary, err error) {

	reviews = AppReviewSummary{} // Needed for marshalling into type

	err = helpers.Unmarshal([]byte(app.Reviews), &reviews)
	log.Err(err)
	return reviews, err
}

func (app App) GetGenreIDs() (genres []int, err error) {

	err = helpers.Unmarshal([]byte(app.Genres), &genres)

	// Needed for marshalling into array
	if len(genres) == 0 {
		genres = []int{}
	}

	return genres, err
}

func (app App) GetGenres() (genres []Genre, err error) {

	var item = helpers.MemcacheAppGenres(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &genres, func() (interface{}, error) {

		ids, err := app.GetGenreIDs()
		if err != nil {
			return genres, err
		}

		return GetGenresByID(ids, []string{"id", "name"})
	})

	if len(genres) == 0 {
		genres = []Genre{} // Needed for marshalling into type
	}

	return genres, err
}

func (app App) GetCategoryIDs() (categories []int, err error) {

	err = helpers.Unmarshal([]byte(app.Categories), &categories)

	// Needed for marshalling into array
	if len(categories) == 0 {
		categories = []int{}
	}

	return categories, err
}

func (app App) GetCategories() (categories []Category, err error) {

	var item = helpers.MemcacheAppCategories(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &categories, func() (interface{}, error) {

		ids, err := app.GetCategoryIDs()
		if err != nil {
			return categories, err
		}

		return GetCategoriesByID(ids, []string{"id", "name"})
	})

	if len(categories) == 0 {
		categories = []Category{} // Needed for marshalling into type
	}

	return categories, err
}

func (app App) GetTagIDs() (tags []int, err error) {

	tags = []int{} // Needed for marshalling into type

	if app.Tags == "" || app.Tags == "null" || app.Tags == "[]" {
		return tags, err
	}

	err = helpers.Unmarshal([]byte(app.Tags), &tags)
	if err != nil {
		log.Err(err)
		return
	}
	return tags, err
}

func (app App) GetTags() (tags []Tag, err error) {

	var item = helpers.MemcacheAppTags(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &tags, func() (interface{}, error) {

		ids, err := app.GetTagIDs()
		if err != nil {
			return tags, err
		}

		return GetTagsByID(ids, []string{"id", "name"})
	})

	if len(tags) == 0 {
		tags = []Tag{} // Needed for marshalling into type
	}

	return tags, err
}

func (app App) GetDeveloperIDs() (developers []int, err error) {

	err = helpers.Unmarshal([]byte(app.Developers), &developers)

	// Needed for marshalling into array
	if len(developers) == 0 {
		developers = []int{}
	}

	return developers, err
}

func (app App) GetDevelopers() (developers []Developer, err error) {

	var item = helpers.MemcacheAppDevelopers(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &developers, func() (interface{}, error) {

		ids, err := app.GetDeveloperIDs()
		if err != nil {
			return developers, err
		}

		return GetDevelopersByID(ids, []string{"id", "name"})
	})

	if len(developers) == 0 {
		developers = []Developer{} // Needed for marshalling into type
	}

	return developers, err
}

func (app App) GetPublisherIDs() (publishers []int, err error) {

	publishers = []int{} // Needed for marshalling into type

	err = helpers.Unmarshal([]byte(app.Publishers), &publishers)
	log.Err(err)
	return publishers, err
}

func (app App) GetPublishers() (publishers []Publisher, err error) {

	var item = helpers.MemcacheAppPublishers(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &publishers, func() (interface{}, error) {

		ids, err := app.GetPublisherIDs()
		if err != nil {
			return publishers, err
		}

		return GetPublishersByID(ids, []string{"id", "name"})
	})

	if len(publishers) == 0 {
		publishers = []Publisher{} // Needed for marshalling into type
	}

	return publishers, err
}

func (app App) GetBundles() (bundles []Bundle, err error) {

	var item = helpers.MemcacheAppBundles(app.ID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &bundles, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return bundles, err
		}

		var bundles []Bundle

		db = db.Where("JSON_CONTAINS(app_ids, '[" + strconv.Itoa(app.ID) + "]')")
		db = db.Find(&bundles)

		return bundles, db.Error
	})

	if len(bundles) == 0 {
		bundles = []Bundle{} // Needed for marshalling into type
	}

	return bundles, err
}

func (app App) GetName() (name string) {
	return helpers.GetAppName(app.ID, app.Name)
}

func (app App) GetMetaImage() string {

	ss, err := app.GetScreenshots()
	if err != nil || len(ss) == 0 {
		return app.GetHeaderImage()
	}
	return ss[0].PathFull
}

func PopularApps() (apps []App, err error) {

	var item = helpers.MemcachePopularApps

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return apps, err
		}

		db = db.Select([]string{"id", "name", "player_peak_week", "background"})
		db = db.Where("type = ?", "game")
		db = db.Order("player_peak_week desc")
		db = db.Limit(30)
		db = db.Find(&apps)

		return apps, err
	})

	return apps, err
}

func PopularNewApps() (apps []App, err error) {

	var item = helpers.MemcachePopularNewApps

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return apps, err
		}

		db = db.Select([]string{"id", "name", "player_peak_week"})
		db = db.Where("type = ?", "game")
		db = db.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix())
		db = db.Order("player_peak_week desc")
		db = db.Limit(25)
		db = db.Find(&apps)

		return apps, err
	})

	return apps, err
}

func TrendingApps() (apps []App, err error) {

	var item = helpers.MemcacheTrendingApps

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return apps, err
		}

		db = db.Select([]string{"id", "name", "player_trend"})
		db = db.Order("player_trend desc")
		db = db.Limit(10)
		db = db.Find(&apps)

		return apps, err
	})

	return apps, err
}

type SteamSpyAppResponse struct {
	Appid     int    `json:"appid"`
	Name      string `json:"name"`
	Developer string `json:"developer"`
	Publisher string `json:"publisher"`
	// ScoreRank      int    `json:"score_rank"` // Can be empty string
	Positive       int    `json:"positive"`
	Negative       int    `json:"negative"`
	Userscore      int    `json:"userscore"`
	Owners         string `json:"owners"`
	AverageForever int    `json:"average_forever"`
	Average2Weeks  int    `json:"average_2weeks"`
	MedianForever  int    `json:"median_forever"`
	Median2Weeks   int    `json:"median_2weeks"`
	Price          string `json:"price"`
	Initialprice   string `json:"initialprice"`
	Discount       string `json:"discount"`
	Languages      string `json:"languages"`
	Genre          string `json:"genre"`
	Ccu            int    `json:"ccu"`
	// Tags           map[string]int `json:"tags"` // Can be an empty slice
}

func (a SteamSpyAppResponse) GetOwners() (ret []int) {

	owners := strings.Replace(a.Owners, ",", "", -1)
	owners = strings.Replace(owners, " ", "", -1)
	ownersStrings := strings.Split(owners, "..")
	ownersInts := helpers.StringSliceToIntSlice(ownersStrings)
	if len(ownersInts) == 2 {
		return ownersInts
	}
	return ret
}

func GetTypesForSelect() []AppType {

	types := []string{
		"game",
		"advertising",
		"application",
		"config",
		"demo",
		"dlc",
		"episode",
		"guide",
		"hardware",
		"media",
		"mod",
		"movie",
		"series",
		"tool",
		"", // Displays as Unknown
		"video",
	}

	var ret []AppType
	for _, v := range types {
		ret = append(ret, AppType{
			ID:   v,
			Name: App{Type: v}.GetType(),
		})
	}

	return ret
}

type AppType struct {
	ID   string
	Name string
}

func GetApp(id int, columns []string) (app App, err error) {

	if id == 0 {
		id = 753
	}

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidAppID
	}

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	if columns != nil && len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return app, db.Error
		}
	}

	db = db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, ErrRecordNotFound
	}

	return app, nil
}

func SearchApp(s string, columns []string) (app App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return app, db.Error
		}
	}

	i, _ := strconv.Atoi(s)
	if helpers.IsValidAppID(i) {
		db = db.First(&app, "id = ?", s)
	} else {
		db = db.First(&app, "name = ?", s)
	}

	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, ErrRecordNotFound
	}

	return app, nil
}

func GetAppsByID(ids []int, columns []string) (apps []App, err error) {

	if len(ids) == 0 {
		return apps, nil
	}

	ids = helpers.Unique(ids)

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db.Where("id IN (?)", ids).Find(&apps)

	return apps, db.Error
}

func GetAppsWithColumnDepth(column string, depth int, columns []string) (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select(columns)
	db = db.Where("JSON_DEPTH("+column+") = ?", depth)
	db = db.Order("id asc")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil

}

func CountApps() (count int, err error) {

	var item = helpers.MemcacheAppsCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Count(&count)

		return count, db.Error
	})

	return count, err
}

func CountAppsWithAchievements() (count int, err error) {

	var item = helpers.MemcacheAppsWithAchievementsCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Where("achievements_count > 0").Count(&count)

		return count, db.Error
	})

	return count, err
}

//
type AppImage struct {
	PathFull      string `json:"f"`
	PathThumbnail string `json:"t"`
}

type AppVideo struct {
	PathFull      string `json:"f"`
	PathThumbnail string `json:"s"`
	Title         string `json:"t"`
}

type AppAchievement struct {
	Name        string  `json:"n"`
	Icon        string  `json:"i"`
	Description string  `json:"d"`
	Completed   float64 `json:"c"`
	Active      bool    `json:"a"`
}

func (a AppAchievement) GetIcon() string {
	if strings.HasSuffix(a.Icon, ".jpg") {
		return a.Icon
	}
	return helpers.DefaultAppIcon
}

type AppStat struct {
	Name        string `json:"n"`
	Default     int    `json:"d"`
	DisplayName string `json:"o"`
}

type AppSteamSpy struct {
	SSAveragePlaytimeTwoWeeks int `json:"aw"`
	SSAveragePlaytimeForever  int `json:"af"`
	SSMedianPlaytimeTwoWeeks  int `json:"mw"`
	SSMedianPlaytimeForever   int `json:"mf"`
	SSOwnersLow               int `json:"ol"`
	SSOwnersHigh              int `json:"oh"`
}

func (ss AppSteamSpy) GetSSAveragePlaytimeTwoWeeks() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSAveragePlaytimeTwoWeeks) / 60)
}

func (ss AppSteamSpy) GetSSAveragePlaytimeForever() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSAveragePlaytimeForever) / 60)
}

func (ss AppSteamSpy) GetSSMedianPlaytimeTwoWeeks() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSMedianPlaytimeTwoWeeks) / 60)
}

func (ss AppSteamSpy) GetSSMedianPlaytimeForever() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSMedianPlaytimeForever) / 60)
}

type AppReviewSummary struct {
	Positive int
	Negative int
	Reviews  []AppReview
}

func (r AppReviewSummary) GetTotal() int {
	return r.Negative + r.Positive
}

func (r AppReviewSummary) GetPositivePercent() float64 {
	return float64(r.Positive) / float64(r.GetTotal()) * 100
}

func (r AppReviewSummary) GetNegativePercent() float64 {
	return float64(r.Negative) / float64(r.GetTotal()) * 100
}

type AppReview struct {
	Review     string `json:"r"`
	Vote       bool   `json:"v"`
	VotesGood  int    `json:"g"`
	VotesFunny int    `json:"f"`
	Created    string `json:"c"`
	PlayerPath string `json:"p"`
	PlayerName string `json:"n"`
}

func (ar AppReview) HTML() template.HTML {
	return template.HTML(ar.Review)
}
