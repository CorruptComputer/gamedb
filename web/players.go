package web

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi"
	slugify "github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/steam"
)

func RanksHandler(w http.ResponseWriter, r *http.Request) {

	// Normalise the order
	var ranks []datastore.Rank
	var err error

	switch chi.URLParam(r, "id") {
	case "badges":
		ranks, err = datastore.GetRanksBy("badges_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].BadgesRank)
		}
	case "friends":
		ranks, err = datastore.GetRanksBy("friends_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].FriendsRank)
		}
	case "games":
		ranks, err = datastore.GetRanksBy("games_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].GamesRank)
		}
	case "level", "":
		ranks, err = datastore.GetRanksBy("level_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].LevelRank)
		}
	case "time":
		ranks, err = datastore.GetRanksBy("play_time_rank")

		for k := range ranks {
			ranks[k].Rank = humanize.Ordinal(ranks[k].PlayTimeRank)
		}
	default:
		err = errors.New("incorrect sort")
	}

	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	// Count players
	playersCount, err := datastore.CountPlayers()
	if err != nil {
		logger.Error(err)
	}

	// Count ranks
	ranksCount, err := datastore.GetRanksCount()
	if err != nil {
		logger.Error(err)
	}

	template := playersTemplate{}
	template.Fill(r, "Ranks")
	template.Ranks = ranks
	template.PlayersCount = playersCount
	template.RanksCount = ranksCount

	returnTemplate(w, r, "ranks", template)
	return
}

type playersTemplate struct {
	GlobalTemplate
	Ranks        []datastore.Rank
	PlayersCount int
	RanksCount   int
}

func PlayerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	slug := chi.URLParam(r, "slug")

	idx, err := strconv.Atoi(id)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	player, err := datastore.GetPlayer(idx)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	errs := player.UpdateIfNeeded()
	if len(errs) > 0 {
		for _, v := range errs {

			logger.Error(err)

			// API is probably down
			if v.Error() == steam.ErrInvalidJson {
				returnErrorTemplate(w, r, 500, "Couldnt fetch player data, steam API may be down?")
				return
			}

			returnErrorTemplate(w, r, 500, err.Error())
			return
		}
	}

	// Redirect to correct slug
	correctSLug := slugify.Make(player.PersonaName)
	if slug != "" && slug != correctSLug {
		http.Redirect(w, r, "/players/"+id+"/"+correctSLug, 302)
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func(player *datastore.Player) {

		// Queue friends
		if player.ShouldUpdateFriends() {

			for _, v := range player.Friends {
				vv, _ := strconv.Atoi(v.SteamID)
				p, _ := json.Marshal(queue.PlayerMessage{
					PlayerID: vv,
				})
				queue.Produce(queue.PlayerQueue, p)
			}

			player.FriendsAddedAt = time.Now()
			player.Save()
		}

		wg.Done()

	}(player)

	var friends []datastore.Player
	wg.Add(1)
	go func(player *datastore.Player) {

		// Make friend ID slice
		var friendsSlice []int
		for _, v := range player.Friends {
			s, _ := strconv.Atoi(v.SteamID)
			friendsSlice = append(friendsSlice, s)
		}

		// Get friends for template
		friends, err = datastore.GetPlayersByIDs(friendsSlice)
		if err != nil {
			logger.Error(err)
		}

		sort.Slice(friends, func(i, j int) bool {
			return friends[i].Level > friends[j].Level
		})

		wg.Done()

	}(player)

	var sortedGamesSlice []*playerAppTemplate
	wg.Add(1)
	go func(player *datastore.Player) {

		// Get games
		var gamesSlice []int
		gamesMap := make(map[int]*playerAppTemplate)
		for _, v := range player.GetGames() {
			gamesSlice = append(gamesSlice, v.AppID)
			gamesMap[v.AppID] = &playerAppTemplate{
				Time: v.PlaytimeForever,
			}
		}

		gamesSql, err := mysql.GetApps(gamesSlice, []string{"id", "name", "price_initial", "icon"})
		if err != nil {
			logger.Error(err)
		}

		for _, v := range gamesSql {
			gamesMap[v.ID].ID = v.ID
			gamesMap[v.ID].Name = v.GetName()
			gamesMap[v.ID].Price = v.GetPriceInitial()
			gamesMap[v.ID].Icon = v.GetIcon()
		}

		// Sort games
		for _, v := range gamesMap {
			sortedGamesSlice = append(sortedGamesSlice, v)
		}

		sort.Slice(sortedGamesSlice, func(i, j int) bool {
			if sortedGamesSlice[i].Time == sortedGamesSlice[j].Time {
				return sortedGamesSlice[i].Name < sortedGamesSlice[j].Name
			}
			return sortedGamesSlice[i].Time > sortedGamesSlice[j].Time
		})

		wg.Done()

	}(player)

	var ranks *datastore.Rank
	wg.Add(1)
	go func(player *datastore.Player) {

		// Get ranks
		ranks, err = datastore.GetRank(player.PlayerID)
		if err != nil {
			if err.Error() != datastore.ErrorNotFound {
				logger.Error(err)
			}
		}

		wg.Done()

	}(player)

	wg.Add(1)
	go func(player *datastore.Player) {

		// Badges
		sort.Slice(player.Badges.Badges, func(i, j int) bool {
			return player.Badges.Badges[i].CompletionTime > player.Badges.Badges[j].CompletionTime
		})

		wg.Done()

	}(player)

	var players int
	wg.Add(1)
	go func(player *datastore.Player) {

		// Number of players
		players, err = datastore.CountPlayers()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}(player)

	// Wait
	wg.Wait()

	// Template
	template := playerTemplate{}
	template.Fill(r, player.PersonaName)
	template.Player = player
	template.Friends = friends
	template.Games = sortedGamesSlice
	template.Ranks = playerRanksTemplate{*ranks, players}

	returnTemplate(w, r, "player", template)
}

type playerTemplate struct {
	GlobalTemplate
	Player  *datastore.Player
	Friends []datastore.Player
	Games   []*playerAppTemplate
	Ranks   playerRanksTemplate
}

type playerAppTemplate struct {
	ID    int
	Name  string
	Price string
	Icon  string
	Time  int
}

func (g playerAppTemplate) GetPriceHour() string {

	price, err := strconv.ParseFloat(g.Price, 64)
	if err != nil {
		price = 0
	}

	x := float64(price) / (float64(g.Time) / 60)
	if math.IsNaN(x) {
		x = 0
	}
	if math.IsInf(x, 0) {
		return "∞"
	}
	return helpers.DollarsFloat(x)
}

type playerRanksTemplate struct {
	Ranks   datastore.Rank
	Players int
}

func (p playerRanksTemplate) format(rank int) string {

	ord := humanize.Ordinal(rank)
	if ord == "0th" {
		return "-"
	}
	return ord
}

func (p playerRanksTemplate) GetLevel() string {
	return p.format(p.Ranks.LevelRank)
}

func (p playerRanksTemplate) GetGames() string {
	return p.format(p.Ranks.GamesRank)
}

func (p playerRanksTemplate) GetBadges() string {
	return p.format(p.Ranks.BadgesRank)
}

func (p playerRanksTemplate) GetTime() string {
	return p.format(p.Ranks.PlayTimeRank)
}

func (p playerRanksTemplate) GetFriends() string {
	return p.format(p.Ranks.FriendsRank)
}

func (p playerRanksTemplate) formatPercent(rank int) string {

	if rank == 0 {
		return ""
	}

	precision := 0
	if rank <= 10 {
		precision = 3
	} else if rank <= 100 {
		precision = 2
	} else if rank <= 1000 {
		precision = 1
	}

	percent := (float64(rank) / float64(p.Players)) * 100
	return strconv.FormatFloat(percent, 'f', precision, 64) + "%"

}

func (p playerRanksTemplate) GetLevelPercent() string {
	return p.formatPercent(p.Ranks.LevelRank)
}

func (p playerRanksTemplate) GetGamesPercent() string {
	return p.formatPercent(p.Ranks.GamesRank)
}

func (p playerRanksTemplate) GetBadgesPercent() string {
	return p.formatPercent(p.Ranks.BadgesRank)
}

func (p playerRanksTemplate) GetTimePercent() string {
	return p.formatPercent(p.Ranks.PlayTimeRank)
}

func (p playerRanksTemplate) GetFriendsPercent() string {
	return p.formatPercent(p.Ranks.FriendsRank)
}

func PlayerIDHandler(w http.ResponseWriter, r *http.Request) {

	post := r.PostFormValue("id")
	post = path.Base(post)

	// Check datastore
	dbPlayer, err := datastore.GetPlayerByName(post)
	if err != nil {

		if err.Error() != datastore.ErrorNotFound {
			logger.Error(err)
		}

		// Check steam
		id, err := steam.GetID(post)
		if err != nil {

			if err != steam.ErrNoUserFound {
				logger.Error(err)
			}

			returnErrorTemplate(w, r, 404, "Can't find user: "+post)
			return
		}

		http.Redirect(w, r, "/players/"+id, 302)
		return
	}

	http.Redirect(w, r, "/players/"+strconv.Itoa(dbPlayer.PlayerID), 302)
	return
}
