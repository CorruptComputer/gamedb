package helpers

import (
	"strconv"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/website/pkg/config"
)

var ErrCacheMiss = memcache.ErrCacheMiss

var memcacheClient = memcache.New("game-db-", config.Config.MemcacheDSN.Get())

func GetMemcache() *memcache.Memcache {
	return memcacheClient
}

var (
	// Counts
	MemcacheAppsCount             = memcache.Item{Key: "apps-Count", Expiration: 86400}
	MemcachePackagesCount         = memcache.Item{Key: "packages-Count", Expiration: 86400}
	MemcacheBundlesCount          = memcache.Item{Key: "bundles-Count", Expiration: 86400}
	MemcacheUpcomingAppsCount     = memcache.Item{Key: "upcoming-apps-Count", Expiration: 86400}
	MemcacheTrendingAppsCount     = memcache.Item{Key: "upcoming-apps-Count", Expiration: 86400}
	MemcacheNewReleaseAppsCount   = memcache.Item{Key: "upcoming-apps-Count", Expiration: 86400}
	MemcacheUpcomingPackagesCount = memcache.Item{Key: "upcoming-packages-Count", Expiration: 86400}
	MemcachePlayersCount          = memcache.Item{Key: "players-Count", Expiration: 86400 * 1}
	MemcachePricesCount           = memcache.Item{Key: "prices-Count", Expiration: 86400 * 7}
	MemcacheMongoCount            = func(key string) memcache.Item {
		return memcache.Item{Key: "mongo-Count-" + key, Expiration: 60 * 60}
	}
	MemcachePlayerEventsCount = func(playerID int64) memcache.Item {
		return memcache.Item{Key: "players-events-Count-" + strconv.FormatInt(playerID, 10), Expiration: 86400}
	}

	// Dropdowns
	MemcacheTagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 86400 * 7}

	// Rows
	MemcacheChangeRow = func(changeID int64) memcache.Item {
		return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 86400 * 30}
	}
	MemcacheConfigRow = func(key string) memcache.Item {
		return memcache.Item{Key: "config-item-" + key, Expiration: 0}
	}
	MemcacheAppPlayersRow = func(appID int) memcache.Item {
		return memcache.Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60}
	}

	// Other
	MemcacheQueues      = memcache.Item{Key: "queues", Expiration: 10}
	MemcachePopularApps = memcache.Item{Key: "popular-apps", Expiration: 60 * 3}
)
