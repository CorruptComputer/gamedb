module github.com/gamedb/gamedb

go 1.13

require (
	cloud.google.com/go/logging v1.1.1
	github.com/Jleagle/go-durationfmt v0.0.0-20190307132420-e57bfad84057
	github.com/Jleagle/influxql v0.0.0-20200804190929-88324f67bffe
	github.com/Jleagle/patreon-go v0.0.0-20201006180837-366bfaa6710a
	github.com/Jleagle/rabbit-go v0.0.0-20200831220529-13e96c303e94
	github.com/Jleagle/recaptcha-go v0.0.0-20200117124940-d00b2c62c076
	github.com/Jleagle/session-go v0.0.0-20190515070633-3c8712426233
	github.com/Jleagle/sitemap-go v0.0.0-20190405195207-2bdddbb3bd50
	github.com/Jleagle/steam-go v0.0.0-20200902113949-8ea84a7fe3f4
	github.com/Jleagle/unmarshal-go v0.0.0-20200217225147-fd7db71d9ac0
	github.com/Philipp15b/go-steam v1.0.1-0.20190816133340-b04c5a83c1c0
	github.com/PuerkitoBio/goquery v1.6.0 // indirect
	github.com/StefanSchroeder/Golang-Roman v0.0.0-20191231161654-ef19f7247884
	github.com/ahmdrz/goinsta/v2 v2.4.5
	github.com/antchfx/xmlquery v1.3.3 // indirect
	github.com/aws/aws-sdk-go v1.35.14 // indirect
	github.com/badoux/checkmail v1.2.1
	github.com/blend/go-sdk v1.1.1 // indirect
	github.com/bwmarrin/discordgo v0.22.0
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/deepmap/oapi-codegen v1.3.13
	github.com/derekstavis/go-qs v0.0.0-20180720192143-9eef69e6c4e7
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.6.0
	github.com/digitalocean/godo v1.49.0
	github.com/djherbis/fscache v0.10.1
	github.com/dustin/go-humanize v1.0.0
	github.com/frustra/bbcode v0.0.0-20180807171629-48be21ce690c
	github.com/getkin/kin-openapi v0.23.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gocolly/colly/v2 v2.1.0
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.2 // indirect
	github.com/google/go-github/v32 v32.1.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/gorilla/websocket v1.4.2
	github.com/gosimple/slug v1.9.0
	github.com/hetznercloud/hcloud-go v1.22.0
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/now v1.1.1
	github.com/justinas/nosurf v1.1.1
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.11.1 // indirect
	github.com/lib/pq v1.3.0 // indirect
	github.com/mailjet/mailjet-apiv3-go v0.0.0-20201009050126-c24bc15a9394
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/mborgerson/GoTruncateHtml v0.0.0-20150507032438-125d9154cd1e
	github.com/memcachier/mc/v3 v3.0.2
	github.com/microcosm-cc/bluemonday v1.0.4
	github.com/montanaflynn/stats v0.6.3
	github.com/mssola/user_agent v0.5.2
	github.com/nicklaw5/helix v1.0.0
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/olivere/elastic/v7 v7.0.21
	github.com/oschwald/maxminddb-golang v1.7.0
	github.com/pariz/gountries v0.0.0-20200430155801-1c6a393df9c7
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.6.2+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.7.0+incompatible
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/streadway/amqp v1.0.0
	github.com/tdewolff/minify/v2 v2.9.10
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/wcharczuk/go-chart v2.0.1+incompatible
	github.com/xdg/stringprep v1.0.0 // indirect
	github.com/yohcop/openid-go v1.0.0
	go.mongodb.org/mongo-driver v1.4.2
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/ratelimit v0.1.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/exp v0.0.0-20200331195152-e8c3332aa8e5 // indirect
	golang.org/x/image v0.0.0-20200927104501-e162460cd6b5 // indirect
	golang.org/x/net v0.0.0-20201024042810-be3efd7ff127
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sys v0.0.0-20201022201747-fb209a7c41cd // indirect
	golang.org/x/text v0.3.3
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.0.0-20201023174141-c8cfbd0f21e6 // indirect
	gonum.org/v1/gonum v0.8.1
	google.golang.org/api v0.33.0
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201022181438-0ff5f38871d5 // indirect
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/djherbis/atime.v1 v1.0.0 // indirect
	gopkg.in/djherbis/stream.v1 v1.3.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
	jaytaylor.com/html2text v0.0.0-20200412013138-3577fbdbcff7
)
