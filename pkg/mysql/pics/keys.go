package pics

import (
	"bytes"
	"html/template"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	picsTypeBool uint8 = iota + 1
	picsTypeBytes
	picsTypeCustom
	picsTypeImage
	picsTypeJSON
	picsTypeLink
	picsTypeMap
	picsTypeNumber
	picsTypeNumberListJSON     // From JSON object
	picsTypeNumberListJSONKeys // From JSON object keys
	picsTypeNumberListString   // From comma string
	picsTypeStringListJSON     // From JSON bject
	picsTypeTextListString     // From comma string
	picsTypeTimestamp
	picsTypeTitle
	picsTypeTooLong
	picsTypePercent
)

var (
	appIDLinkType = PicsKey{FormatType: picsTypeLink, Link: "/games/$val$"}
)

var CommonKeys = map[string]PicsKey{
	"allowpurchasefromrestrictedcountries": {FormatType: picsTypeBool},
	"app_retired_publisher_request":        {FormatType: picsTypeBool},
	"associations":                         {FormatType: picsTypeCustom},
	"category":                             {FormatType: picsTypeCustom, Link: "/games?categories=$val$"},
	"clienticns":                           {FormatType: picsTypeLink, Link: helpers.AppIconBase + "$app$/$val$.icns"},
	"clienticon":                           {FormatType: picsTypeImage, Link: helpers.AppIconBase + "$app$/$val$.ico"},
	"clienttga":                            {FormatType: picsTypeLink, Link: helpers.AppIconBase + "$app$/$val$.tga"},
	"community_hub_visible":                {FormatType: picsTypeBool},
	"community_visible_stats":              {FormatType: picsTypeBool},
	"controller_support":                   {FormatType: picsTypeTitle},
	"controllervr":                         {FormatType: picsTypeNumberListJSONKeys},
	"disableoverlay":                       {FormatType: picsTypeBool},
	"eulas":                                {FormatType: picsTypeCustom},
	"exfgls":                               {FormatType: picsTypeBool, Description: "Exclude from game library sharing"},
	"freeondemand":                         {FormatType: picsTypeBool},
	"gameid":                               appIDLinkType,
	"genres":                               {FormatType: picsTypeNumberListJSON, Link: "/games?genres=$val$"},
	"has_adult_content":                    {FormatType: picsTypeBool},
	"header_image":                         {FormatType: picsTypeMap},
	"icon":                                 {FormatType: picsTypeImage, Link: helpers.AppIconBase + "$app$/$val$.jpg"},
	"isplugin":                             {FormatType: picsTypeBool},
	"languages":                            {FormatType: picsTypeCustom},
	"library_assets":                       {FormatType: picsTypeJSON},
	"linuxclienticon":                      {FormatType: picsTypeLink, Link: helpers.AppIconBase + "$app$/$val$.zip"},
	"logo":                                 {FormatType: picsTypeImage, Link: helpers.AppIconBase + "$app$/$val$.jpg"},
	"logo_small":                           {FormatType: picsTypeImage, Link: helpers.AppIconBase + "$app$/$val$.jpg"},
	"market_presence":                      {FormatType: picsTypeBool},
	"metacritic_fullurl":                   {FormatType: picsTypeLink, Link: "$val$"},
	"metacritic_score":                     {FormatType: picsTypePercent},
	"name_localized":                       {FormatType: picsTypeStringListJSON},
	"onlyvrsupport":                        {FormatType: picsTypeBool},
	"openvr_controller_bindings":           {FormatType: picsTypeJSON},
	"openvrsupport":                        {FormatType: picsTypeBool},
	"original_release_date":                {FormatType: picsTypeTimestamp},
	"oslist":                               {FormatType: picsTypeTextListString, Link: "/games?platforms=$val$"},
	"osvrsupport":                          {FormatType: picsTypeBool},
	"parent":                               appIDLinkType,
	"parentappid":                          appIDLinkType,
	"playareavr":                           {FormatType: picsTypeJSON},
	"primary_genre":                        {FormatType: picsTypeLink, Link: "/games?genres=$val$"},
	"releasestate":                         {FormatType: picsTypeTitle},
	"releasestateoverrideinverse":          {FormatType: picsTypeBool},
	"requireskbmouse":                      {FormatType: picsTypeBool},
	"review_percentage":                    {FormatType: picsTypePercent},
	"service_app":                          {FormatType: picsTypeBool},
	"small_capsule":                        {FormatType: picsTypeMap},
	"steam_release_date":                   {FormatType: picsTypeTimestamp},
	"steamchinaapproved":                   {FormatType: picsTypeBool},
	"store_asset_mtime":                    {FormatType: picsTypeTimestamp},
	"store_tags":                           {FormatType: picsTypeNumberListJSON, Link: "/games?tags=$val$"},
	"supported_languages":                  {FormatType: picsTypeCustom},
	"systemprofile":                        {FormatType: picsTypeBool},
	"type":                                 {FormatType: picsTypeTitle},
	"visibleonlyonavailableplatforms":      {FormatType: picsTypeBool},
	"workshop_visible":                     {FormatType: picsTypeBool},
}

var ExtendedKeys = map[string]PicsKey{
	"aaalaunchredirect":                    appIDLinkType,
	"absolutemousecoordinates":             {FormatType: picsTypeBool},
	"allowconversion":                      {FormatType: picsTypeBool},
	"allowcrossregiontradingandgifting":    {FormatType: picsTypeBool},
	"allowelevation":                       {FormatType: picsTypeBool},
	"allowpurchasefromrestrictedcountries": {FormatType: picsTypeBool},
	"anti_cheat_support_url":               {FormatType: picsTypeLink, Link: "$val$"},
	"betaforappid":                         appIDLinkType,
	"canskipinstallappchooser":             {FormatType: picsTypeBool},
	"checkpkgstate":                        {FormatType: picsTypeBool},
	"curatorconnect":                       {FormatType: picsTypeBool},
	"demoofappid	":                      appIDLinkType,
	"dependantonapp":                       appIDLinkType,
	"dependantonapppreventrecurseholes":    {FormatType: picsTypeBool},
	"developer_url":                        {FormatType: picsTypeLink, Link: "$val$"},
	"disable_shader_precaching":            {FormatType: picsTypeBool},
	"disablemanifestiteration":             {FormatType: picsTypeBool},
	"disableosxdrmloader":                  {FormatType: picsTypeBool},
	"disableoverlay":                       {FormatType: picsTypeBool},
	"disableoverlay_linux":                 {FormatType: picsTypeBool},
	"disableoverlay_macos":                 {FormatType: picsTypeBool},
	"disableoverlay_windows":               {FormatType: picsTypeBool},
	"disableoverlayinjection":              {FormatType: picsTypeBool},
	"disablesendclosesignal":               {FormatType: picsTypeBool},
	"disablestreaming":                     {FormatType: picsTypeBool},
	"dlcavailableonstore":                  {FormatType: picsTypeBool},
	"dlcforappid":                          appIDLinkType,
	"dlcpurchasefromingame":                {FormatType: picsTypeBool},
	"encrypted_video":                      {FormatType: picsTypeBool},
	"expansionofappid":                     appIDLinkType,
	"externallyupdated":                    {FormatType: picsTypeBool},
	"forcelaunchoptions":                   {FormatType: picsTypeBool},
	"gamemanualurl":                        {FormatType: picsTypeLink, Link: "$val$"},
	"guideappid":                           appIDLinkType,
	"hadthirdpartycdkey":                   {FormatType: picsTypeBool},
	"hasexternalregistrationurl":           {FormatType: picsTypeBool},
	"hdaddon":                              appIDLinkType,
	"homepage":                             {FormatType: picsTypeLink, Link: "$val$"},
	"ignorechildprocesses":                 {FormatType: picsTypeBool},
	"inhibitautoversionroll":               {FormatType: picsTypeBool},
	"isconverteddlc":                       {FormatType: picsTypeBool},
	"isfreeapp":                            {FormatType: picsTypeBool},
	"ismediafile":                          {FormatType: picsTypeBool},
	"languages":                            {FormatType: picsTypeTextListString},
	"languages_macos":                      {FormatType: picsTypeTextListString},
	"launchredirect":                       appIDLinkType,
	"legacykeylinkedexternally":            {FormatType: picsTypeBool},
	"listofdlc":                            {FormatType: picsTypeNumberListString, Link: "/games/$val$"},
	"loadallbeforelaunch":                  {FormatType: picsTypeBool},
	"manage_steamguard_useweb":             {FormatType: picsTypeBool},
	"musicalbumavailableonstore":           {FormatType: picsTypeBool},
	"musicalbumforappid":                   appIDLinkType,
	"mustownapptopurchase":                 appIDLinkType,
	"no_revenue_accumlation":               {FormatType: picsTypeBool},
	"no_revenue_accumulation":              {FormatType: picsTypeBool},
	"nodefaultenglishcontent":              {FormatType: picsTypeBool},
	"noservers":                            {FormatType: picsTypeBool},
	"onlyallowrestrictedcountries":         {FormatType: picsTypeBool},
	"optionaldlc":                          {FormatType: picsTypeBool},
	"overlaymanuallyclearscreen":           {FormatType: picsTypeBool},
	"purchasedisabledreason":               {FormatType: picsTypeBool},
	"requiredappid":                        appIDLinkType,
	"requirentfspartition":                 {FormatType: picsTypeBool},
	"requiressse":                          {FormatType: picsTypeBool},
	"retailautostart":                      {FormatType: picsTypeBool},
	"sdk_notownedbydefault":                {FormatType: picsTypeBool},
	"showcdkeyinmenu":                      {FormatType: picsTypeBool},
	"showcdkeyonlaunch":                    {FormatType: picsTypeBool},
	"sourcegame":                           {FormatType: picsTypeBool},
	"supports64bit":                        {FormatType: picsTypeBool},
	"supportscdkeycopytoclipboard":         {FormatType: picsTypeBool},
	"suppressims":                          {FormatType: picsTypeBool},
	"thirdpartycdkey":                      {FormatType: picsTypeBool},
	"vacmacmodulecache":                    appIDLinkType,
	"vacmodulecache":                       appIDLinkType,
	"validoslist":                          {FormatType: picsTypeTextListString, Link: "/games?platforms=$val$"},
	"visibleonlywheninstalled":             {FormatType: picsTypeBool},
	"visibleonlywhensubscribed":            {FormatType: picsTypeBool},
	"vrheadsetstreaming":                   {FormatType: picsTypeBool},
}

var ConfigKeys = map[string]PicsKey{
	"cegpublickey":                       {FormatType: picsTypeTooLong},
	"checkforupdatesbeforelaunch":        {FormatType: picsTypeBool},
	"duration_control_show_interstitial": {FormatType: picsTypeBool},
	"enable_duration_control":            {FormatType: picsTypeBool},
	"enabletextfiltering":                {FormatType: picsTypeBool},
	"gameoverlay_testmode":               {FormatType: picsTypeBool},
	"installscriptoverride":              {FormatType: picsTypeBool},
	"installscriptsignature":             {FormatType: picsTypeTooLong},
	"launchwithoutworkshopupdates":       {FormatType: picsTypeBool},
	"matchmaking_uptodate":               {FormatType: picsTypeBool},
	"noupdatesafterinstall":              {FormatType: picsTypeBool},
	"signaturescheckedonlaunch":          {FormatType: picsTypeJSON},
	"signedfiles":                        {FormatType: picsTypeJSON},
	"steamcontrollerconfigdetails":       {FormatType: picsTypeJSON},
	"steamcontrollertemplateindex":       {FormatType: picsTypeBool},
	"steamcontrollertouchconfigdetails":  {FormatType: picsTypeJSON},
	"steamcontrollertouchtemplateindex":  {FormatType: picsTypeBool},
	"systemprofile":                      {FormatType: picsTypeBool},
	"uselaunchcommandline":               {FormatType: picsTypeBool},
	"usemms":                             {FormatType: picsTypeBool},
	"usesfrenemies":                      {FormatType: picsTypeBool},
	"verifyupdates":                      {FormatType: picsTypeBool},
	"vrcompositorsupport":                {FormatType: picsTypeBool},
}

var UFSKeys = map[string]PicsKey{
	"hidecloudui":         {FormatType: picsTypeBool},
	"ignoreexternalfiles": {FormatType: picsTypeBool},
	"maxnumfiles":         {FormatType: picsTypeNumber},
	"quota":               {FormatType: picsTypeBytes},
	"rootoverrides":       {FormatType: picsTypeJSON},
	"savefiles":           {FormatType: picsTypeCustom},
}

type PicsKey struct {
	FormatType  uint8
	Link        string
	Description string
}

func getType(key string, keys map[string]PicsKey) uint8 {

	if val, ok := keys[key]; ok {
		return val.FormatType
	}
	return 0
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

			return outputHTML("link.gohtml", map[string]interface{}{
				"Link":  item.Link,
				"Blank": !strings.HasPrefix(item.Link, "/"),
				"Val":   val,
			})

		case picsTypeImage:

			if val == "" {
				return ""
			}

			item.Link = strings.ReplaceAll(item.Link, "$val$", val)
			item.Link = strings.ReplaceAll(item.Link, "$app$", strconv.Itoa(appID))

			return template.HTML("<a href=\"" + item.Link + "\" target=\"_blank\" rel=\"noopener\"><img class=\"wide\" data-lazy=\"" + item.Link + "\" alt=\"\" data-lazy-alt=\"" + key + "\" /></a>")

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

			// Shrink keys
			j = regexp.MustCompile("([A-Z0-9]{31,})").ReplaceAllStringFunc(j, func(s string) string {
				return s[0:30] + "..."
			})

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
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"noopener\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeTitle:

			return strings.Title(val)

		case picsTypeNumberListJSON:

			idMap := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &idMap)
			if err != nil {
				log.ErrS(err, val)
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
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"noopener\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeStringListJSON:

			m := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &m)
			if err != nil {
				log.ErrS(err, val)
			}

			var items []string
			for k, v := range m {
				items = append(items, "<li>"+k+": <span class=font-weight-bold>"+v+"</span></li>")
			}

			return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")

		case picsTypeNumberListJSONKeys:

			idMap := map[string]string{}

			err := helpers.Unmarshal([]byte(val), &idMap)
			if err != nil {
				log.ErrS(err, val)
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
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"noopener\">" + id + "</a>"
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
					idSlice[k] = "<a href=\"" + strings.ReplaceAll(item.Link, "$val$", id) + "\" rel=\"noopener\">" + id + "</a>"
				}
			}

			return template.HTML(strings.Join(idSlice, ", "))

		case picsTypeTooLong:

			val = regexp.MustCompile("([a-zA-Z0-9]{31,})").ReplaceAllStringFunc(val, func(s string) string {
				return s[0:30] + "..."
			})

		case picsTypeMap:

			if val != "" {

				m := map[string]string{}
				err := helpers.Unmarshal([]byte(val), &m)
				if err != nil {
					log.ErrS(err, val)
				}

				var items []string
				for k, v := range m {
					items = append(items, "<li>"+k+": <span class=font-weight-bold>"+v+"</span></li>")
				}

				return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
			}

		case picsTypePercent:

			return val + "%"

		case picsTypeCustom:

			switch key {
			case "eulas":

				if val != "" {

					eulas := EULAs{}
					_ = helpers.Unmarshal([]byte(val), &eulas)

					var items []string
					for _, eula := range eulas {
						if eula.Name == "" {
							eula.Name = "EULA"
						}
						items = append(items, `<li><a target="_blank" rel="noopener" href="`+eula.URL+`">`+string(eula.Name)+`</a></li>`)
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "supported_languages":

				if val != "" {

					langs := SupportedLanguages{}
					err := helpers.Unmarshal([]byte(val), &langs)
					if err != nil {
						log.ErrS(err, val)
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
						log.ErrS(err, val)
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

					var itemStrings []string
					if item.Link != "" {
						for _, v := range items {
							id := strconv.Itoa(v)
							link := strings.ReplaceAll(item.Link, "$val$", id)
							itemStrings = append(itemStrings, "<a href=\""+link+"\" rel=\"noopener\">"+id+"</a>")
						}
					}

					return template.HTML(strings.Join(itemStrings, ", "))
				}

			case "languages":

				if val != "" {

					languages := map[string]string{}
					err := helpers.Unmarshal([]byte(val), &languages)
					if err != nil {
						log.ErrS(err, val)
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
						log.ErrS(err, val)
					}

					var items []string
					for _, v := range associations {
						items = append(items, "<li>"+strings.Title(v.Type)+": <span class=font-weight-bold>"+v.Name+"</span></li>")
					}

					return template.HTML("<ul class='mb-0 pl-3'>" + strings.Join(items, "") + "</ul>")
				}

			case "savefiles":

				if val == "" {
					return ""
				}

				files := saveFiles{}
				err := helpers.Unmarshal([]byte(val), &files)
				if err != nil {
					log.ErrS(err, val)
				}

				for k := range files {
					if files[k].Path == "{}" {
						files[k].Path = ""
					}
				}

				return outputHTML("savefiles.gohtml", map[string]interface{}{
					"Files": files,
				})
			}

			return val
		}
	}

	return val
}

func outputHTML(filename string, data interface{}) template.HTML {

	t, err := template.ParseFiles("./templates/pics_keys/" + filename)
	if err != nil {
		log.ErrS(err)
		return ""
	}

	b := bytes.NewBufferString("")

	if err := t.Execute(b, data); err != nil {
		log.ErrS(err)
		return ""
	}

	return template.HTML(b.String())
}
