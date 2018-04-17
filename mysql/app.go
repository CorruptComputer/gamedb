package mysql

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosimple/slug"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
)

type App struct {
	ID                     int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`                  //
	CreatedAt              *time.Time `gorm:"not null;column:created_at"`                                     //
	UpdatedAt              *time.Time `gorm:"not null;column:updated_at"`                                     //
	AchievementPercentages string     `gorm:"not null;column:achievement_percentages;type:text;default:'[]'"` // JSON
	Achievements           string     `gorm:"not null;column:achievements;type:text;default:'{}'"`            // JSON
	Background             string     `gorm:"not null;column:background"`                                     //
	Categories             string     `gorm:"not null;column:categories;type:json;default:'[]'"`              // JSON
	ChangeNumber           int        `gorm:"not null;column:change_number"`                                  // PICS
	ClientIcon             string     `gorm:"not null;column:client_icon"`                                    // PICS
	ComingSoon             bool       `gorm:"not null;column:coming_soon"`                                    //
	Developers             string     `gorm:"not null;column:developers;type:json;default:'[]'"`              // JSON
	DLC                    string     `gorm:"not null;column:dlc;type:json;default:'[]'"`                     // JSON
	Extended               string     `gorm:"not null;column:extended;default:'{}'"`                          // JSON
	GameID                 int        `gorm:"not null;column:game_id"`                                        //
	GameName               string     `gorm:"not null;column:game_name"`                                      //
	Genres                 string     `gorm:"not null;column:genres;type:json;default:'[]'"`                  // JSON
	Ghost                  bool       `gorm:"not null;column:is_ghost;type:tinyint(1)"`                       //
	HeaderImage            string     `gorm:"not null;column:image_header"`                                   //
	Homepage               string     `gorm:"not null;column:homepage"`                                       // PICS
	Icon                   string     `gorm:"not null;column:icon"`                                           // PICS
	IsFree                 bool       `gorm:"not null;column:is_free;type:tinyint(1)"`                        //
	Logo                   string     `gorm:"not null;column:logo"`                                           // PICS
	MetacriticScore        int8       `gorm:"not null;column:metacritic_score"`                               //
	MetacriticURL          string     `gorm:"not null;column:metacritic_url"`                                 //
	Movies                 string     `gorm:"not null;column:movies;type:text;default:'[]'"`                  // JSON
	Name                   string     `gorm:"not null;column:name"`                                           //
	Packages               string     `gorm:"not null;column:packages;type:json;default:'[]'"`                // JSON
	Platforms              string     `gorm:"not null;column:platforms;type:json;default:'[]'"`               // JSON
	PriceDiscount          int        `gorm:"not null;column:price_discount"`                                 //
	PriceFinal             int        `gorm:"not null;column:price_final"`                                    //
	PriceInitial           int        `gorm:"not null;column:price_initial"`                                  //
	Publishers             string     `gorm:"not null;column:publishers;type:json;default:'[]'"`              // JSON
	ReleaseDate            string     `gorm:"not null;column:release_date"`                                   //
	ReleaseState           string     `gorm:"not null;column:release_state"`                                  // PICS
	Schema                 string     `gorm:"not null;column:schema;type:text;default:'{}'"`                  // JSON
	Screenshots            string     `gorm:"not null;column:screenshots;type:text;default:'[]'"`             // JSON
	ShortDescription       string     `gorm:"not null;column:description_short"`                              //
	StoreTags              string     `gorm:"not null;column:tags;type:json;default:'[]'"`                    // PICS JSON
	Type                   string     `gorm:"not null;column:type"`                                           //
}

func GetDefaultAppJSON() App {
	return App{
		AchievementPercentages: "[]",
		Achievements:           "{}",
		Categories:             "[]",
		Developers:             "[]",
		DLC:                    "[]",
		Extended:               "{}",
		Genres:                 "[]",
		Movies:                 "[]",
		Packages:               "[]",
		Platforms:              "[]",
		Publishers:             "[]",
		Schema:                 "{}",
		Screenshots:            "[]",
		StoreTags:              "[]",
	}
}

func (app App) GetPath() (ret string) {
	ret = "/games/" + strconv.Itoa(app.ID)

	if app.Name != "" {
		ret = ret + "/" + slug.Make(app.Name)
	}

	return ret
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

func (app App) GetDefaultAvatar() string {
	return "/assets/img/no-app-image-square.jpg"
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

	return helpers.GetReleaseDateNice(app.ReleaseDate)
}

func (app App) GetReleaseDateUnix() int64 {

	return helpers.GetReleaseDateUnix(app.ReleaseDate)
}

func (app App) GetIcon() (ret string) {

	if app.Icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(app.ID) + "/" + app.Icon + ".jpg"
	}
}

func (app App) GetPriceInitial() float64 {
	return helpers.CentsInt(app.PriceInitial)
}

func (app App) GetCommunityLink() (string) {
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=SteamAuthority&utm_medium=link&utm_campaign=SteamAuthority"
}

func (app App) GetStoreLink() (string) {
	return "https://store.steampowered.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=SteamAuthority&utm_medium=link&utm_campaign=SteamAuthority"
}

func (app App) GetPCGamingWikiLink() (string) {
	return "https://pcgamingwiki.com/api/appid.php?appid=" + strconv.Itoa(app.ID)
}

func (app App) GetInstallLink() (template.URL) {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetMetacriticLink() (template.URL) {
	return template.URL("http://www.metacritic.com/game/" + app.MetacriticURL)
}

// Used in frontend
func (app App) GetScreenshots() (screenshots []steam.AppDetailsScreenshot, err error) {

	bytes := []byte(app.Screenshots)
	if err := json.Unmarshal(bytes, &screenshots); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return screenshots, err
	}

	return screenshots, nil
}

// Used in template
func (app App) GetCoopTags(tagMap map[int]string) (string) {

	tags, err := app.GetTags()
	if err != nil {
		logger.Error(err)
	}

	var coopTags []string
	for _, v := range tags {
		if val, ok := tagMap[v]; ok {
			coopTags = append(coopTags, val)
		}
	}

	return strings.Join(coopTags, ", ")
}

func (app App) GetAchievements() (achievements steam.AppDetailsAchievements, err error) {

	bytes := []byte(app.Achievements)
	if err := json.Unmarshal(bytes, &achievements); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return achievements, err
	}

	return achievements, nil
}

func (app App) GetPlatforms() (platforms []string, err error) {

	bytes := []byte(app.Platforms)
	if len(bytes) == 0 {
		return platforms, nil
	}

	if err := json.Unmarshal(bytes, &platforms); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return platforms, err
	}

	return platforms, nil
}

func (app App) GetPlatformImages() (ret template.HTML, err error) {

	platforms, err := app.GetPlatforms()
	if err != nil {
		return ret, err
	}

	for _, v := range platforms {
		if v == "macos" {
			ret = ret + `<i class="fab fa-apple"></i>`
		} else if v == "windows" {
			ret = ret + `<i class="fab fa-windows"></i>`
		} else if v == "linux" {
			ret = ret + `<i class="fab fa-linux"></i>`
		}
	}

	return ret, nil
}

func (app App) GetDLC() (dlcs []int, err error) {

	bytes := []byte(app.DLC)
	if err := json.Unmarshal(bytes, &dlcs); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return dlcs, err
	}

	return dlcs, nil
}

func (app App) GetPackages() (packages []int, err error) {

	bytes := []byte(app.Packages)
	if err := json.Unmarshal(bytes, &packages); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return packages, err
	}

	return packages, nil
}

func (app App) GetGenres() (genres []steam.AppDetailsGenre, err error) {

	bytes := []byte(app.Genres)
	if err := json.Unmarshal(bytes, &genres); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return genres, err
	}

	return genres, nil
}

func (app App) GetCategories() (categories []string, err error) {

	bytes := []byte(app.Categories)
	if err := json.Unmarshal(bytes, &categories); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return categories, err
	}

	return categories, nil
}

func (app App) GetTags() (tags []int, err error) {

	bytes := []byte(app.StoreTags)
	if err := json.Unmarshal(bytes, &tags); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return tags, err
	}

	return tags, nil
}

func (app App) GetDevelopers() (developers []string, err error) {

	bytes := []byte(app.Developers)
	if err := json.Unmarshal(bytes, &developers); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return developers, err
	}

	return developers, nil
}

func (app App) GetPublishers() (publishers []string, err error) {

	bytes := []byte(app.Publishers)
	if err := json.Unmarshal(bytes, &publishers); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		}
		return publishers, err
	}

	return publishers, nil
}

func (app App) GetName() (name string) {

	if app.Name == "" {
		app.Name = "App " + strconv.Itoa(app.ID)
	}

	return app.Name
}

func GetApp(id int) (app App, err error) {

	db, err := GetDB()
	if err != nil {
		return app, err
	}

	db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, errors.New("no id")
	}

	return app, nil
}

func GetApps(ids []int, columns []string) (apps []App, err error) {

	if len(ids) < 1 {
		return apps, nil
	}

	db, err := GetDB()
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

func SearchApps(query url.Values, limit int, sort string, columns []string) (apps []App, err error) {

	db, err := GetDB()
	if err != nil {
		return apps, err
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if sort != "" {
		db = db.Order(sort)
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	// Hide ghosts
	db = db.Where("is_ghost = ?", 0)

	// Type
	if _, ok := query["type"]; ok {
		db = db.Where("type = ?", query.Get("type"))
	}

	// Tags depth
	if _, ok := query["tags_depth"]; ok {
		db = db.Where("JSON_DEPTH(tags) = ?", query.Get("tags_depth"))
	}

	// Genres depth
	if _, ok := query["genres_depth"]; ok {
		db = db.Where("JSON_DEPTH(genres) = ?", query.Get("genres_depth"))
	}

	// Free
	if _, ok := query["is_free"]; ok {
		db = db.Where("is_free = ?", query.Get("is_free"))
	}

	// Platforms
	if _, ok := query["platforms"]; ok {
		db = db.Where("JSON_CONTAINS(platforms, [\"?\"])", query.Get("platforms"))

	}

	// Tag
	if _, ok := query["tags"]; ok {
		db = db.Where("JSON_CONTAINS(tags, ?)", "[\""+query.Get("tags")+"\"]")
	}

	// Genres
	// select * from apps WHERE JSON_SEARCH(genres, 'one', 'Action') IS NOT NULL;

	// Query
	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, err
}

func GetDLC(app App, columns []string) (apps []App, err error) {

	db, err := GetDB()
	if err != nil {
		return apps, err
	}

	dlc, err := app.GetDLC()
	if err != nil {
		return apps, err
	}

	db = db.Where("id in (?)", dlc).Find(&apps)

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func CountApps() (count int, err error) {

	db, err := GetDB()
	if err != nil {
		return count, err
	}

	db.Model(&App{}).Count(&count)
	if db.Error != nil {
		return count, db.Error
	}

	return count, nil
}

func (app *App) Update() (errs []error) {

	var wg sync.WaitGroup

	// Get details from store API
	wg.Add(1)
	go func(app *App) {

		response, err := steam.GetAppDetailsFromStore(app.ID)
		if err != nil {

			if err == steam.ErrGhostApp {
				app.Ghost = true
			}

			if err == steam.ErrBadResponse {
				errs = append(errs, err)
			}
		}

		// Screenshots
		screenshotsString, err := json.Marshal(response.Data.Screenshots)
		if err != nil {
			errs = append(errs, err)
		}

		// Movies
		moviesString, err := json.Marshal(response.Data.Movies)
		if err != nil {
			errs = append(errs, err)
		}

		// Achievements
		achievementsString, err := json.Marshal(response.Data.Achievements)
		if err != nil {
			errs = append(errs, err)
		}

		// DLC
		dlcString, err := json.Marshal(response.Data.DLC)
		if err != nil {
			errs = append(errs, err)
		}

		// Packages
		packagesString, err := json.Marshal(response.Data.Packages)
		if err != nil {
			errs = append(errs, err)
		}

		// Publishers
		publishersString, err := json.Marshal(response.Data.Publishers)
		if err != nil {
			errs = append(errs, err)
		}

		// Developers
		developersString, err := json.Marshal(response.Data.Developers)
		if err != nil {
			errs = append(errs, err)
		}

		// Categories
		var categories []int8
		for _, v := range response.Data.Categories {
			categories = append(categories, v.ID)
		}

		categoriesString, err := json.Marshal(categories)
		if err != nil {
			errs = append(errs, err)
		}

		genresString, err := json.Marshal(response.Data.Genres)
		if err != nil {
			errs = append(errs, err)
		}

		// Platforms
		var platforms []string
		if response.Data.Platforms.Linux {
			platforms = append(platforms, "linux")
		}
		if response.Data.Platforms.Windows {
			platforms = append(platforms, "windows")
		}
		if response.Data.Platforms.Windows {
			platforms = append(platforms, "macos")
		}

		platformsString, err := json.Marshal(platforms)
		if err != nil {
			errs = append(errs, err)
		}

		// Other
		app.Name = response.Data.Name
		app.Type = response.Data.Type
		app.IsFree = response.Data.IsFree
		app.DLC = string(dlcString)
		app.ShortDescription = response.Data.ShortDescription
		app.HeaderImage = response.Data.HeaderImage
		app.Developers = string(developersString)
		app.Publishers = string(publishersString)
		app.Packages = string(packagesString)
		app.MetacriticScore = response.Data.Metacritic.Score
		app.MetacriticURL = response.Data.Metacritic.URL
		app.Categories = string(categoriesString)
		app.Genres = string(genresString)
		app.Screenshots = string(screenshotsString)
		app.Movies = string(moviesString)
		app.Achievements = string(achievementsString)
		app.Background = response.Data.Background
		app.Platforms = string(platformsString)
		app.GameID = response.Data.Fullgame.AppID
		app.GameName = response.Data.Fullgame.Name
		app.ReleaseDate = response.Data.ReleaseDate.Date
		app.ComingSoon = response.Data.ReleaseDate.ComingSoon
		app.PriceInitial = response.Data.PriceOverview.Initial
		app.PriceFinal = response.Data.PriceOverview.Final
		app.PriceDiscount = response.Data.PriceOverview.DiscountPercent

		wg.Done()
	}(app)

	// Get summary
	wg.Add(1)
	go func(app *App) {

		response, err := steam.GetPICSInfo([]int{app.ID}, []int{})
		if err != nil {
			errs = append(errs, err)
		}

		var js steam.JsApp
		if len(response.Apps) > 0 {
			js = response.Apps[strconv.Itoa(app.ID)]
		} else {
			errs = append(errs, errors.New("no app key in json"))
		}

		// Check if empty
		app.Ghost = reflect.DeepEqual(js.Common, steam.JsAppCommon{})

		// Tags, convert map to slice
		var tagsSlice []int
		for _, v := range js.Common.StoreTags {
			vv, _ := strconv.Atoi(v)
			tagsSlice = append(tagsSlice, vv)
		}

		tags, err := json.Marshal(tagsSlice)
		if err != nil {
			errs = append(errs, err)
		}

		// Meta critic
		var metacriticScoreInt = 0
		if js.Common.MetacriticScore != "" {
			metacriticScoreInt, _ = strconv.Atoi(js.Common.MetacriticScore)
		}

		// Extended
		extended, err := json.Marshal(js.Extended)
		if err != nil {
			errs = append(errs, err)
		}

		//
		app.Name = js.Common.Name
		app.Type = js.Common.Type
		app.ReleaseState = js.Common.ReleaseState
		// app.Platforms = strings.Split(js.Common.OSList, ",") // Can get from API
		app.MetacriticScore = int8(metacriticScoreInt)
		app.MetacriticURL = js.Common.MetacriticURL
		app.StoreTags = string(tags)
		// app.Developers = js.Extended.Developer // Store API can handle multiple values
		// app.Publishers = js.Extended.Publisher // Store API can handle multiple values
		app.Homepage = js.Extended.Homepage
		app.ChangeNumber = js.ChangeNumber
		app.Logo = js.Common.Logo
		app.Icon = js.Common.Icon
		app.ClientIcon = js.Common.ClientIcon
		app.Extended = string(extended)

		wg.Done()
	}(app)

	// Achievement percentages
	wg.Add(1)
	go func(app *App) {

		percentages, err := steam.GetGlobalAchievementPercentagesForApp(app.ID)
		if err != nil {
			logger.Error(err)
		}

		percentagesString, err := json.Marshal(percentages)
		if err != nil {
			logger.Error(err)
		}

		app.AchievementPercentages = string(percentagesString)

		wg.Done()
	}(app)

	// Schema
	wg.Add(1)
	go func(app *App) {

		schema, err := steam.GetSchemaForGame(app.ID)
		if err != nil {
			logger.Error(err)
		}

		schemaString, err := json.Marshal(schema)
		if err != nil {
			logger.Error(err)
		}

		app.Schema = string(schemaString)

		wg.Done()
	}(app)

	// Wait
	wg.Wait()

	// Tidy
	app.Type = strings.ToLower(app.Type)
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	// Default JSON values
	if app.StoreTags == "" || app.StoreTags == "null" {
		app.StoreTags = "[]"
	}

	if app.Categories == "" || app.Categories == "null" {
		app.Categories = "[]"
	}

	if app.Genres == "" || app.Genres == "null" {
		app.Genres = "[]"
	}

	if app.Screenshots == "" || app.Screenshots == "null" {
		app.Screenshots = "[]"
	}

	if app.Movies == "" || app.Movies == "null" {
		app.Movies = "[]"
	}

	if app.Achievements == "" || app.Achievements == "null" {
		app.Achievements = "{}"
	}

	if app.Platforms == "" || app.Platforms == "null" {
		app.Platforms = "[]"
	}

	if app.DLC == "" || app.DLC == "null" {
		app.DLC = "[]"
	}

	if app.Packages == "" || app.Packages == "null" {
		app.Packages = "[]"
	}

	if app.AchievementPercentages == "" || app.AchievementPercentages == "null" {
		app.AchievementPercentages = "[]"
	}

	if app.Developers == "" || app.Developers == "null" {
		app.Developers = "[]"
	}

	if app.Publishers == "" || app.Publishers == "null" {
		app.Publishers = "[]"
	}

	if app.Schema == "" || app.Schema == "null" {
		app.Schema = "{}"
	}

	if app.Extended == "" || app.Extended == "null" {
		app.Extended = "{}"
	}

	return errs
}
