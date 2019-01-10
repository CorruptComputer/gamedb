package db

import (
	"encoding/json"
	"errors"
	"html/template"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"

	DefaultAppIcon = "/assets/img/no-app-image-square.jpg"
)

type App struct {
	ID                     int       `gorm:"not null;column:id;primary_key"`                   //
	CreatedAt              time.Time `gorm:"not null;column:created_at;type:datetime"`         //
	UpdatedAt              time.Time `gorm:"not null;column:updated_at;type:datetime"`         //
	PICSChangeNumber       int       `gorm:"not null;column:change_number"`                    //
	PICSChangeNumberDate   time.Time `gorm:"not null;column:change_number_date;type:datetime"` //
	PICSCommon             string    `gorm:"not null;column:common"`                           //
	PICSConfig             string    `gorm:"not null;column:config"`                           //
	PICSDepots             string    `gorm:"not null;column:depots"`                           //
	PICSExtended           string    `gorm:"not null;column:extended"`                         //
	PICSInstall            string    `gorm:"not null;column:install"`                          //
	PICSLaunch             string    `gorm:"not null;column:launch"`                           //
	PICSLocalization       string    `gorm:"not null;column:localization"`                     //
	PICSPublicOnly         bool      `gorm:"not null;column:public_only"`                      //
	PICSSystemRequirements string    `gorm:"not null;column:system_requirements"`              //
	PICSUFS                string    `gorm:"not null;column:ufs"`                              //
	Achievements           string    `gorm:"not null;column:achievements;type:text"`           // []AppAchievement
	Background             string    `gorm:"not null;column:background"`                       //
	BundleIDs              string    `gorm:"not null;column:bundle_ids"`                       //
	Categories             string    `gorm:"not null;column:categories;type:json"`             //
	ClientIcon             string    `gorm:"not null;column:client_icon"`                      //
	ComingSoon             bool      `gorm:"not null;column:coming_soon"`                      //
	Developers             string    `gorm:"not null;column:developers;type:json"`             //
	DLC                    string    `gorm:"not null;column:dlc;type:json"`                    //
	DLCCount               int       `gorm:"not null;column:dlc_count"`                        //
	GameID                 int       `gorm:"not null;column:game_id"`                          //
	GameName               string    `gorm:"not null;column:game_name"`                        //
	Genres                 string    `gorm:"not null;column:genres;type:json"`                 //
	HeaderImage            string    `gorm:"not null;column:image_header"`                     //
	Homepage               string    `gorm:"not null;column:homepage"`                         //
	Icon                   string    `gorm:"not null;column:icon"`                             //
	IsFree                 bool      `gorm:"not null;column:is_free;type:tinyint(1)"`          //
	Logo                   string    `gorm:"not null;column:logo"`                             //
	MetacriticScore        int8      `gorm:"not null;column:metacritic_score"`                 //
	MetacriticURL          string    `gorm:"not null;column:metacritic_url"`                   //
	Movies                 string    `gorm:"not null;column:movies;type:text"`                 // []AppVideo
	Name                   string    `gorm:"not null;column:name"`                             //
	NewsIDs                string    `gorm:"not null;column:news_ids"`                         //
	Packages               string    `gorm:"not null;column:packages;type:json"`               // []int
	Platforms              string    `gorm:"not null;column:platforms;type:json"`              //
	Prices                 string    `gorm:"not null;column:prices"`                           //
	Publishers             string    `gorm:"not null;column:publishers;type:json"`             //
	ReleaseDate            string    `gorm:"not null;column:release_date"`                     //
	ReleaseDateUnix        int64     `gorm:"not null;column:release_date_unix"`                //
	ReleaseState           string    `gorm:"not null;column:release_state"`                    //
	Reviews                string    `gorm:"not null;column:reviews"`                          //
	ReviewsNegative        int       `gorm:"not null;column:reviews_negative"`                 //
	ReviewsPositive        int       `gorm:"not null;column:reviews_positive"`                 //
	ReviewsScore           float64   `gorm:"not null;column:reviews_score"`                    //
	Screenshots            string    `gorm:"not null;column:screenshots;type:text"`            // []AppImage
	ShortDescription       string    `gorm:"not null;column:description_short"`                //
	Stats                  string    `gorm:"not null;column:stats;type:text"`                  // []AppStat
	SteamSpy               string    `gorm:"not null;column:steam_spy"`                        // AppSteamSpy
	StoreTags              string    `gorm:"not null;column:tags;type:json"`                   //
	Type                   string    `gorm:"not null;column:type"`                             //
	Version                string    `gorm:"not null;column:version"`                          //

	SSAveragePlaytimeTwoWeeks int `gorm:"not null;column:ss_average_2weeks"`
	SSAveragePlaytimeForever  int `gorm:"not null;column:ss_average_forever"`
	SSMedianPlaytimeForever   int `gorm:"not null;column:ss_median_forever"`
	SSMedianPlaytimeTwoWeeks  int `gorm:"not null;column:ss_median_2weeks"`
	SSOwnersLow               int `gorm:"not null;column:ss_owners_low"`
	SSOwnersHigh              int `gorm:"not null;column:ss_owners_high"`
}

func (app *App) BeforeCreate(scope *gorm.Scope) error {

	if app.Achievements == "" {
		app.Achievements = "[]"
	}
	if app.Categories == "" {
		app.Categories = "[]"
	}
	if app.Developers == "" {
		app.Developers = "[]"
	}
	if app.DLC == "" {
		app.DLC = "[]"
	}
	if app.PICSExtended == "" {
		app.PICSExtended = "{}"
	}
	if app.PICSSystemRequirements == "" {
		app.PICSSystemRequirements = "{}"
	}
	if app.Prices == "" {
		app.Prices = "{}"
	}
	if app.Genres == "" {
		app.Genres = "[]"
	}
	if app.Movies == "" {
		app.Movies = "[]"
	}
	if app.Packages == "" {
		app.Packages = "[]"
	}
	if app.Platforms == "" {
		app.Platforms = "[]"
	}
	if app.Publishers == "" {
		app.Publishers = "[]"
	}
	if app.Stats == "" {
		app.Stats = "[]"
	}
	if app.Screenshots == "" {
		app.Screenshots = "[]"
	}
	if app.StoreTags == "" {
		app.StoreTags = "[]"
	}

	return nil
}

func (app App) GetID() int {
	return app.ID
}

func (app App) GetProductType() ProductType {
	return ProductTypeApp
}

func (app App) GetPath() string {
	return GetAppPath(app.ID, app.Name)
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

func (app App) GetDaysToRelease() string {

	return helpers.GetDaysToRelease(app.ReleaseDateUnix)
}

func (app App) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		app.ReviewsScore,
		GetPriceFormatted(app, code).Final,
		app.PICSChangeNumberDate.Unix(),
	}
}

// Must be the same as package OutputForJSONUpcoming
func (app App) OutputForJSONUpcoming(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		GetPriceFormatted(app, code).Final,
		app.GetDaysToRelease(),
		app.GetReleaseDateNice(),
	}
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

	d := app.PICSChangeNumberDate

	// Empty dates
	if d.IsZero() || d.Unix() == -62167219200 {
		return "-"
	}
	return d.Format(helpers.DateYearTime)
}

func (app App) GetIcon() (ret string) {
	return GetAppIcon(app.ID, app.Icon)
}

func (app *App) SetPrices(prices ProductPrices) (err error) {

	bytes, err := json.Marshal(prices)
	if err != nil {
		return err
	}

	app.Prices = string(bytes)

	return nil
}

func (app App) GetPrices() (prices ProductPrices, err error) {

	err = helpers.Unmarshal([]byte(app.Prices), &prices)
	return prices, err
}

func (app App) GetPrice(code steam.CountryCode) (price ProductPriceStruct, err error) {

	prices, err := app.GetPrices()
	if err != nil {
		return price, err
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

// Adds to current news IDs
func (app *App) SetNewsIDs(news steam.News) (err error) {

	ids, err := app.GetNewsIDs()
	if err != nil {
		return err
	}

	for _, v := range news.Items {
		ids = append(ids, v.GID)
	}

	bytes, err := json.Marshal(helpers.Unique64(ids))
	if err != nil {
		return err
	}

	app.NewsIDs = string(bytes)
	return nil
}

func (app *App) SetReviewScore() {

	if app.ReviewsPositive == 0 && app.ReviewsNegative == 0 {

		app.ReviewsScore = 0

	} else {

		total := float64(app.ReviewsPositive + app.ReviewsNegative)
		average := float64(app.ReviewsPositive) / total
		score := average - (average-0.5)*math.Pow(2, -math.Log10(total + 1))

		app.ReviewsScore = helpers.RoundFloatTo2DP(score * 100)
	}
}

func (app *App) SetExtended(extended PICSExtended) (err error) {

	bytes, err := json.Marshal(extended)
	if err != nil {
		return err
	}

	app.PICSExtended = string(bytes)

	return nil
}

func (app App) GetExtended() (extended PICSExtended, err error) {

	extended = PICSExtended{}

	err = helpers.Unmarshal([]byte(app.PICSExtended), &extended)
	return extended, err
}

func (app *App) SetCommon(common PICSAppCommon) (err error) {

	bytes, err := json.Marshal(common)
	if err != nil {
		return err
	}

	app.PICSCommon = string(bytes)

	return nil
}

func (app App) GetCommon() (common PICSAppCommon, err error) {

	common = PICSAppCommon{}

	err = helpers.Unmarshal([]byte(app.PICSCommon), &common)
	return common, err
}

func (app *App) SetConfig(config PICSAppConfig) (err error) {

	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	app.PICSConfig = string(bytes)

	return nil
}

func (app App) GetConfig() (config PICSAppConfig, err error) {

	config = PICSAppConfig{}

	err = helpers.Unmarshal([]byte(app.PICSConfig), &config)
	return config, err
}

func (app *App) SetDepots(depots PicsDepots) (err error) {

	bytes, err := json.Marshal(depots)
	if err != nil {
		return err
	}

	app.PICSDepots = string(bytes)

	return nil
}

func (app App) GetDepots() (depots PicsDepots, err error) {

	err = helpers.Unmarshal([]byte(app.PICSDepots), &depots)
	return depots, err
}

func (app *App) SetLaunch(items []PICSAppConfigLaunchItem) (err error) {

	bytes, err := json.Marshal(items)
	if err != nil {
		return err
	}

	app.PICSLaunch = string(bytes)

	return nil
}

func (app App) GetLaunch() (items []PICSAppConfigLaunchItem, err error) {

	err = helpers.Unmarshal([]byte(app.PICSLaunch), &items)
	return items, err
}

func (app *App) SetInstall(install map[string]interface{}) (err error) {

	bytes, err := json.Marshal(install)
	if err != nil {
		return err
	}

	app.PICSInstall = string(bytes)

	return nil
}

func (app App) GetInstall() (install map[string]interface{}, err error) {

	install = map[string]interface{}{}
	err = helpers.Unmarshal([]byte(app.PICSInstall), &install)
	return install, err
}

func (app *App) SetLocalization(localization map[string]interface{}) (err error) {

	bytes, err := json.Marshal(localization)
	if err != nil {
		return err
	}

	app.PICSLocalization = string(bytes)

	return nil
}

func (app App) GetLocalization() (localization map[string]interface{}, err error) {

	localization = map[string]interface{}{}
	err = helpers.Unmarshal([]byte(app.PICSLocalization), &localization)
	return localization, err
}

func (app *App) SetSystemRequirements(systemRequirements map[string]interface{}) (err error) {

	bytes, err := json.Marshal(systemRequirements)
	if err != nil {
		return err
	}

	app.PICSSystemRequirements = string(bytes)

	return nil
}

func (app App) GetSystemRequirements() (systemRequirements map[string]interface{}, err error) {

	systemRequirements = map[string]interface{}{}
	err = helpers.Unmarshal([]byte(app.PICSSystemRequirements), &systemRequirements)
	return systemRequirements, err
}

func (app *App) SetUFS(ufs PICSAppUFS) (err error) {

	bytes, err := json.Marshal(ufs)
	if err != nil {
		return err
	}

	app.PICSUFS = string(bytes)

	return nil
}

func (app App) GetUFS() (ufs PICSAppUFS, err error) {

	ufs = PICSAppUFS{}
	err = helpers.Unmarshal([]byte(app.PICSUFS), &ufs)
	return ufs, err
}

func (app App) GetCommunityLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (app App) GetStoreLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://store.steampowered.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (app App) GetPCGamingWikiLink() string {
	return "https://pcgamingwiki.com/api/appid.php?appid=" + strconv.Itoa(app.ID)
}

func (app App) GetHeaderImage() string {
	return "http://cdn.akamai.steamstatic.com/steam/apps/" + strconv.Itoa(app.ID) + "/header.jpg"
}

func (app App) GetInstallLink() template.URL {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetMetacriticLink() template.URL {
	return template.URL("http://www.metacritic.com/game/" + app.MetacriticURL)
}

func (app App) GetScreenshots() (screenshots []AppImage, err error) {

	err = helpers.Unmarshal([]byte(app.Screenshots), &screenshots)
	return screenshots, err
}

func (app App) GetMovies() (movies []AppVideo, err error) {

	err = helpers.Unmarshal([]byte(app.Movies), &movies)
	return movies, err
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

func (app App) GetAchievements() (achievements []AppAchievement, err error) {

	err = helpers.Unmarshal([]byte(app.Achievements), &achievements)
	return achievements, err
}

func (app App) GetStats() (stats []AppStat, err error) {

	err = helpers.Unmarshal([]byte(app.Stats), &stats)
	return stats, err
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
		return ret, err
	}

	if helpers.SliceHasString(platforms, platformWindows) {
		ret = ret + `<i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformMac) {
		ret = ret + `<i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformLinux) {
		ret = ret + `<i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret, nil
}

func (app App) GetDLC() (dlcs []int, err error) {

	err = helpers.Unmarshal([]byte(app.DLC), &dlcs)
	return dlcs, err
}

func (app App) GetPackages() (packages []int, err error) {

	err = helpers.Unmarshal([]byte(app.Packages), &packages)
	return packages, err
}

func (app App) GetReviews() (reviews steam.ReviewsResponse, err error) {

	err = helpers.Unmarshal([]byte(app.Reviews), &reviews)
	return reviews, err
}

func (app App) GetGenres() (genres []steam.AppDetailsGenre, err error) {

	err = helpers.Unmarshal([]byte(app.Genres), &genres)
	return genres, err
}

func (app App) GetCategories() (categories []string, err error) {

	err = helpers.Unmarshal([]byte(app.Categories), &categories)
	return categories, err
}

func (app App) GetTagIDs() (tags []int, err error) {

	err = helpers.Unmarshal([]byte(app.StoreTags), &tags)
	return tags, err
}

func (app App) GetTags() (tags []Tag, err error) {

	ids, err := app.GetTagIDs()
	if err != nil {
		return tags, err
	}

	return GetTagsByID(ids)
}

func (app App) GetDevelopers() (developers []string, err error) {

	err = helpers.Unmarshal([]byte(app.Developers), &developers)
	return developers, err
}

func (app App) GetPublishers() (publishers []string, err error) {

	err = helpers.Unmarshal([]byte(app.Publishers), &publishers)
	return publishers, err
}

func (app App) GetName() (name string) {
	return getAppName(app.ID, app.Name)
}

type SteamSpyApp struct {
	Appid     int    `json:"appid"`
	Name      string `json:"name"`
	Developer string `json:"developer"`
	Publisher string `json:"publisher"`
	//ScoreRank      int    `json:"score_rank"` // Can be empty string
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
	//Tags           map[string]int `json:"tags"` // Can be an empty slice
}

func (a SteamSpyApp) GetOwners() (ret []int) {

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

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	db = db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return app, db.Error
		}
	}

	if app.ID == 0 {
		return app, ErrRecordNotFound
	}

	return app, nil
}

func GetAppsByID(ids []int, columns []string) (apps []App, err error) { // todo, chunk ids into multple queries async

	if len(ids) == 0 {
		return apps, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db.Where("id IN (?)", ids).Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

// todo, these methods could all be one?
func GetAppsWithTags() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"tags", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(tags) = 2")
	db = db.Order("id asc")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithPackages() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"packages"})
	db = db.Where("JSON_DEPTH(packages) = 2")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithDevelopers() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"developers", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(developers) = 2")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithPublishers() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"publishers", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(publishers) = 2")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithGenres() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"genres", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(genres) = 3")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetDLC(app App, columns []string) (apps []App, err error) {

	dlc, err := app.GetDLC()
	if err != nil {
		return apps, err
	}

	if len(dlc) == 0 {
		return apps, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Where("id in (?)", dlc).Find(&apps)

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	return apps, db.Error
}

func CountApps() (count int, err error) {

	return helpers.GetMemcache().GetSetInt(helpers.MemcacheAppsCount, func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Count(&count)
		return count, db.Error
	})
}

func GetMostExpensiveApp(code steam.CountryCode) (price int, err error) {

	return helpers.GetMemcache().GetSetInt(helpers.MemcacheMostExpensiveApp(code), func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		var countSlice []int
		db.Model(&App{}).Pluck("max(prices->\"$."+string(code)+".final\")", &countSlice)
		if db.Error != nil {
			return count, db.Error
		}
		if len(countSlice) != 1 {
			return count, errors.New("query failed")
		}

		return countSlice[0], nil
	})
}

func IsValidAppID(id int) bool {
	return id != 0
}

func GetAppPath(id int, name string) string {

	p := "/apps/" + strconv.Itoa(id)

	if name != "" {
		p = p + "/" + slug.Make(name)
	}

	return p
}

func getAppName(id int, name string) string {

	if name != "" {
		return name
	} else if id > 0 {
		return "App " + strconv.Itoa(id)
	}
	return "Unknown App"
}

func GetAppIcon(id int, icon string) string {

	if icon == "" {
		return DefaultAppIcon
	} else if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(id) + "/" + icon + ".jpg"
}

func CountUpcomingApps() (count int, err error) {

	return helpers.GetMemcache().GetSetInt(helpers.MemcacheUpcomingAppsCount, func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db = db.Model(new(App))
		db = db.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		db = db.Count(&count)

		return count, db.Error
	})
}

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

type AppReview struct {
}
