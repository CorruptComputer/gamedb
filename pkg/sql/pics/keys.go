package pics

import (
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type PicsKeyFormatType string

const (
	picsTypeBool               PicsKeyFormatType = "bool"
	picsTypeBytes              PicsKeyFormatType = "bytes"
	picsTypeCustom             PicsKeyFormatType = "custom"
	picsTypeImage              PicsKeyFormatType = "image"
	picsTypeJSON               PicsKeyFormatType = "json"
	picsTypeLink               PicsKeyFormatType = "link"
	picsTypeMap                PicsKeyFormatType = "map"
	picsTypeNumber             PicsKeyFormatType = "number"
	picsTypeNumberListJSON     PicsKeyFormatType = "number-list-json"      // From JSON object
	picsTypeNumberListJSONKeys PicsKeyFormatType = "number-list-json-keys" // From JSON object keys
	picsTypeNumberListString   PicsKeyFormatType = "number-list-string"    // From comma string
	picsTypeTextListString     PicsKeyFormatType = "text-list-string"      // From comma string
	picsTypeTimestamp          PicsKeyFormatType = "timestamp"
	picsTypeTitle              PicsKeyFormatType = "title"
)

var CommonKeys = map[string]PicsKey{
	"app_retired_publisher_request": {FormatType: picsTypeBool},
	"associations":                  {FormatType: picsTypeCustom},
	"category":                      {FormatType: picsTypeCustom},
	"clienticns":                    {FormatType: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.icns"},
	"clienticon":                    {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.ico"},
	"clienttga":                     {FormatType: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.tga"},
	"community_hub_visible":         {FormatType: picsTypeBool},
	"community_visible_stats":       {FormatType: picsTypeBool},
	"controller_support":            {FormatType: picsTypeTitle},
	"controllervr":                  {FormatType: picsTypeNumberListJSONKeys},
	"eulas":                         {FormatType: picsTypeCustom},
	"exfgls":                        {FormatType: picsTypeBool, Description: "Exclude from game library sharing"},
	"gameid":                        {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"genres":                        {FormatType: picsTypeNumberListJSON, Link: "/apps?genres=$val$"},
	"has_adult_content":             {FormatType: picsTypeBool},
	"header_image":                  {FormatType: picsTypeMap},
	"icon":                          {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"languages":                     {FormatType: picsTypeCustom},
	"library_assets":                {FormatType: picsTypeJSON},
	"linuxclienticon":               {FormatType: picsTypeLink, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.zip"},
	"logo":                          {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"logo_small":                    {FormatType: picsTypeImage, Link: "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/$app$/$val$.jpg"},
	"metacritic_fullurl":            {FormatType: picsTypeLink, Link: "$val$"},
	"metacritic_score":              {FormatType: picsTypeCustom},
	"onlyvrsupport":                 {FormatType: picsTypeBool},
	"openvr_controller_bindings":    {FormatType: picsTypeJSON},
	"original_release_date":         {FormatType: picsTypeTimestamp},
	"oslist":                        {FormatType: picsTypeTextListString, Link: "/apps?platforms=$val$"},
	"openvrsupport":                 {FormatType: picsTypeBool},
	"parent":                        {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"playareavr":                    {FormatType: picsTypeJSON},
	"primary_genre":                 {FormatType: picsTypeLink, Link: "/apps?genres=$val$"},
	"releasestate":                  {FormatType: picsTypeTitle},
	"small_capsule":                 {FormatType: picsTypeMap},
	"steam_release_date":            {FormatType: picsTypeTimestamp},
	"store_asset_mtime":             {FormatType: picsTypeTimestamp},
	"store_tags":                    {FormatType: picsTypeNumberListJSON, Link: "/apps?tags=$val$"},
	"supported_languages":           {FormatType: picsTypeCustom},
	"type":                          {FormatType: picsTypeTitle},
	"workshop_visible":              {FormatType: picsTypeBool},
}

var ExtendedKeys = map[string]PicsKey{
	"allowcrossregiontradingandgifting":    {FormatType: picsTypeBool},
	"allowpurchasefromrestrictedcountries": {FormatType: picsTypeBool},
	"anti_cheat_support_url":               {FormatType: picsTypeLink, Link: "$val$"},
	"curatorconnect":                       {FormatType: picsTypeBool},
	"developer_url":                        {FormatType: picsTypeLink, Link: "$val$"},
	"dlcavailableonstore":                  {FormatType: picsTypeBool},
	"gamemanualurl":                        {FormatType: picsTypeLink, Link: "$val$"},
	"homepage":                             {FormatType: picsTypeLink, Link: "$val$"},
	"isconverteddlc":                       {FormatType: picsTypeBool},
	"isfreeapp":                            {FormatType: picsTypeBool},
	"languages":                            {FormatType: picsTypeTextListString},
	"listofdlc":                            {FormatType: picsTypeNumberListString, Link: "/apps/$val$"},
	"loadallbeforelaunch":                  {FormatType: picsTypeBool},
	"musicalbumforappid":                   {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"noservers":                            {FormatType: picsTypeBool},
	"requiressse":                          {FormatType: picsTypeBool},
	"sourcegame":                           {FormatType: picsTypeBool},
	"vacmacmodulecache":                    {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"vacmodulecache":                       {FormatType: picsTypeLink, Link: "/apps/$val$"},
	"validoslist":                          {FormatType: picsTypeTextListString, Link: "/apps?platforms=$val$"},
	"visibleonlywheninstalled":             {FormatType: picsTypeBool},
	"visibleonlywhensubscribed":            {FormatType: picsTypeBool},
	"vrheadsetstreaming":                   {FormatType: picsTypeBool},
	"showcdkeyinmenu":                      {FormatType: picsTypeBool},
	"showcdkeyonlaunch":                    {FormatType: picsTypeBool},
	"supportscdkeycopytoclipboard":         {FormatType: picsTypeBool},
}

var ConfigKeys = map[string]PicsKey{
	"checkforupdatesbeforelaunch":  {FormatType: picsTypeBool},
	"enabletextfiltering":          {FormatType: picsTypeBool},
	"installscriptoverride":        {FormatType: picsTypeBool},
	"launchwithoutworkshopupdates": {FormatType: picsTypeBool},
	"matchmaking_uptodate":         {FormatType: picsTypeBool},
	"signaturescheckedonlaunch":    {FormatType: picsTypeJSON},
	"signedfiles":                  {FormatType: picsTypeJSON},
	"steamcontrollerconfigdetails": {FormatType: picsTypeJSON},
	"steamcontrollertemplateindex": {FormatType: picsTypeBool},
	"systemprofile":                {FormatType: picsTypeBool},
	"usesfrenemies":                {FormatType: picsTypeBool},
	"usemms":                       {FormatType: picsTypeBool},
	"verifyupdates":                {FormatType: picsTypeBool},
	"vrcompositorsupport":          {FormatType: picsTypeBool},
}

var UFSKeys = map[string]PicsKey{
	"hidecloudui":   {FormatType: picsTypeBool},
	"maxnumfiles":   {FormatType: picsTypeNumber},
	"quota":         {FormatType: picsTypeBytes},
	"savefiles":     {FormatType: picsTypeCustom},
	"rootoverrides": {FormatType: picsTypeJSON},
}

type PicsKey struct {
	FormatType  PicsKeyFormatType
	Link        string
	Description string
}

func getType(key string, keys map[string]PicsKey) PicsKeyFormatType {

	if val, ok := keys[key]; ok {
		return val.FormatType
	}
	return ""
}

func getDescription(key string, keys map[string]PicsKey) string {

	if val, ok := keys[key]; ok {
		return val.Description
	}
	return ""
}

func FormatVal(key string, val string, appID int, keys map[string]PicsKey) interface{} {

	if item, ok := keys[key]; ok {
		switch item.FormatType {
		case picsTypeBool:

			b, _ := strconv.ParseBool(val)
			if b || val == "yes" {
				return template.HTML("<i class=\"fas fa-check text-success\"></i>")
			}
			return template.HTML("<i class=\"fas fa-times text-danger\"></i>")

		case picsTypeLink:

			if val == "" {
				return ""
			}

			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))

			var blank string
			if !strings.HasPrefix(item.Link, "/") {
				blank = " rel=\"nofollow\" target=\"_blank\""
			}

			return template.HTML("<a href=\"" + item.Link + "\"" + blank + " rel=\"nofollow\">" + val + "</a>")

		case picsTypeImage:

			if val == "" {
				return ""
			}

			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))

			return template.HTML("<div class=\"icon-name\"><div class=\"icon\"><img class=\"wide\" data-lazy=\"" + item.Link + "\" alt=\"\" data-lazy-alt=\"" + key + "\" /></div><div class=\"name\"><a href=\"" + item.Link + "\" rel=\"nofollow\" target=\"_blank\">" + val + "</a></div></div>")

		case picsTypeTimestamp:

			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return val
			}

			return time.Unix(i, 0).Format(helpers.DateTime)

		case picsTypeBytes:

			i, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return val
			}

			return humanize.Bytes(i)

		case picsTypeNumber:

			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return val
			}

			return humanize.Comma(i)

		case picsTypeJSON:

			j, err := helpers.FormatJSON(val)
			if err != nil {
				return val
			}

			return template.HTML("<div class=\"json\">" + j + "</div>")

		case picsTypeNumberListString:

			var idSlice []string

			ids := strings.Split(val, ",")
			for _, id := range ids {
				id = strings.TrimSpace(id)
				idSlice = append(idSlice, id)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				a, _ := strconv.Atoi(idSlice[i])
				b, _ := strconv.Atoi(idSlice[j])
				return a < b
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeTitle:

			return strings.Title(val)

		case picsTypeNumberListJSON:

			idMap := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &idMap)
			if err != nil {
				log.Err(err, val)
			}

			var idSlice []string

			for _, id := range idMap {
				idSlice = append(idSlice, id)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				a, _ := strconv.Atoi(idSlice[i])
				b, _ := strconv.Atoi(idSlice[j])
				return a < b
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeNumberListJSONKeys:

			idMap := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &idMap)
			if err != nil {
				log.Err(err, val)
			}

			var idSlice []string

			for k := range idMap {
				idSlice = append(idSlice, k)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				return idSlice[i] < idSlice[j]
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return strings.Join(idSlice, ", ")

		case picsTypeTextListString:

			var idSlice []string

			ids := strings.Split(val, ",")
			for _, id := range ids {
				id = strings.TrimSpace(id)
				idSlice = append(idSlice, id)
			}

			sort.Slice(idSlice, func(i, j int) bool {
				return idSlice[i] < idSlice[j]
			})

			if item.Link != "" {
				for k, id := range idSlice {
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"nofollow\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeMap:

			if val != "" {

				m := map[string]string{}
				err := helpers.Unmarshal([]byte(val), &m)
				if err != nil {
					log.Err(err, val)
				}

				var items []string
				for k, v := range m {
					items = append(items, "<li>"+k+": <span class=font-weight-bold>"+v+"</span></li>")
				}

				return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
			}

		case picsTypeCustom:

			switch key {
			case "eulas":

				if val != "" {

					eulas := EULAs{}
					err := helpers.Unmarshal([]byte(val), &eulas)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for _, eula := range eulas {
						items = append(items, `<li><a target="_blank" href="`+eula.URL+`">`+eula.Name+`</a></li>`)
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "supported_languages":

				if val != "" {

					langs := SupportedLanguages{}
					err := helpers.Unmarshal([]byte(val), &langs)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for code, lang := range langs {

						var item = code.Title()
						var features []string

						// if lang.Supported {
						// 	item += " <i class=\"fas fa-check text-success\"></i>"
						// } else {
						// 	item += " <i class=\"fas fa-times text-danger\"></i>"
						// }

						if lang.FullAudio {
							features = append(features, "Full Audio")
						}
						if lang.Subtitles {
							features = append(features, "Subtitles")
						}

						if len(features) > 0 {
							item += " + " + strings.Join(features, ", ")
						}

						items = append(items, item)
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i] < items[j]
					})

					return template.HTML(strings.Join(items, ", "))
				}

			case "category":

				if val != "" {

					categories := map[string]string{}
					err := helpers.Unmarshal([]byte(val), &categories)
					if err != nil {
						log.Err(err, val)
					}

					var items []int
					for k := range categories {

						i, err := strconv.Atoi(strings.Replace(k, "category_", "", 1))
						if err == nil {
							items = append(items, i)
						}
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i] < items[j]
					})

					return helpers.JoinInts(items, ", ")
				}

			case "languages":

				if val != "" {

					languages := map[string]string{}
					err := helpers.Unmarshal([]byte(val), &languages)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for k, v := range languages {
						if v == "1" {
							items = append(items, k)
						}
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i] < items[j]
					})

					return strings.Join(items, ", ")
				}

			case "associations":

				if val != "" {

					associations := Associations{}
					err := helpers.Unmarshal([]byte(val), &associations)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for _, v := range associations {
						items = append(items, "<li>"+strings.Title(v.Type)+": <span class=font-weight-bold>"+v.Name+"</span></li>")
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "metacritic_score":

				return template.HTML(val + "<small>/100</small>")

			case "savefiles":

				if val != "" {

					files := saveFiles{}
					err := helpers.Unmarshal([]byte(val), &files)
					if err != nil {
						log.Err(err, val)
					}

					var items []string
					for _, file := range files {
						items = append(items, `<li><strong>Path: </strong>`+file.Path+`, <strong>Pattern: </strong>`+file.Pattern+`, <strong>Recursive: </strong>`+file.Recursive+`, <strong>Root: </strong>`+file.Root+`</li>`)
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			}

			return val
		}
	}

	return val
}
