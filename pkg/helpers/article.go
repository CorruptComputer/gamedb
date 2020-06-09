package helpers

import (
	"bytes"
	"html/template"
	"regexp"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/net/html"
)

const articleImageBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/clans/"

var trimArticles = regexp.MustCompile(`(?s)^(<br>)+|(<br>)+$`)

func GetArticleBody(body string) template.HTML {

	body = strings.ReplaceAll(body, "{STEAM_CLAN_IMAGE}", articleImageBase)
	body = strings.ReplaceAll(body, "{STEAM_CLAN_LOC_IMAGE}", articleImageBase)

	doc, err := html.Parse(strings.NewReader(body))
	if err == nil {
		updateArticleDom(doc)
		var buf = bytes.NewBufferString("")
		err := html.Render(buf, doc)
		log.Err(err)
		body = buf.String()
	}

	body = BBCodeCompiler.Compile(body)
	body = html.UnescapeString(body)
	body = trimArticles.ReplaceAllString(body, "")

	return template.HTML(body)
}

func updateArticleDom(n *html.Node) {

	// Remove image heights to make responsive
	if n.Type == html.ElementNode && n.Data == "img" {

		i := -1
		for index, attr := range n.Attr {
			if attr.Key == "height" {
				i = index
				break
			}
		}
		if i != -1 {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
		}
	}

	// Set target on links
	if n.Type == html.ElementNode && n.Data == "a" {

		var set bool
		for k, v := range n.Attr {
			if v.Key == "target" {
				n.Attr[k].Val = "_blank"
				set = true
				break
			}
		}

		if !set {
			n.Attr = append(n.Attr, html.Attribute{Namespace: "", Key: "target", Val: "_blank"})
			n.Attr = append(n.Attr, html.Attribute{Namespace: "", Key: "rel", Val: "noopener"})
		}
	}

	// Recurse
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		updateArticleDom(c)
	}
}
