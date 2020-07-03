package helpers

import (
	"bytes"
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/net/html"
)

const articleImageBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/clans/"

var fixBBCodeSpaces = regexp.MustCompile(`\](\s+)\[`)

func GetArticleIcon(articleIcon string, appID int, appIcon string) string {

	if appIcon != "" {
		return GetAppIcon(appID, appIcon)
	}

	if strings.HasPrefix(articleIcon, "http") {

		params := url.Values{}
		params.Set("url", articleIcon)
		params.Set("w", "64")
		params.Set("h", "64")
		params.Set("output", "webp")
		params.Set("t", "square")

		return "https://images.weserv.nl?" + params.Encode()
	}

	return DefaultAppIcon
}

func GetArticleBody(body string) template.HTML {

	body = strings.ReplaceAll(body, "{STEAM_CLAN_IMAGE}", articleImageBase)
	body = strings.ReplaceAll(body, "{STEAM_CLAN_LOC_IMAGE}", articleImageBase)

	body = fixBBCodeSpaces.ReplaceAllString(body, "][") // Remove new lines and spaces inbetween bbcodes
	body = BBCodeCompiler.Compile(body)

	// Fix up HTML
	doc, err := html.Parse(strings.NewReader(body))
	if err == nil {
		updateArticleDom(doc)
		var buf = bytes.NewBufferString("")
		err := html.Render(buf, doc)
		log.Err(err)
		body = buf.String()
	}

	body = html.UnescapeString(body)

	return template.HTML(body)
}

func updateArticleDom(n *html.Node) {

	// Images
	if n.Type == html.ElementNode && n.Data == "img" {

		// Remove image heights to make responsive
		removeAttribute(n, "height")

		// Lazy load
		var src = getAttribute(n, "src")
		if src != "" {
			setAttribute(n, "src", "")
			setAttribute(n, "data-lazy", src)
		}
	}

	// Links
	if n.Type == html.ElementNode && n.Data == "a" {

		var href = getAttribute(n, "href")
		if strings.HasSuffix(href, ".jpg") {
			// Remove links to images
			removeAttribute(n, "href")
			removeAttribute(n, "target")
		} else {
			// Open links in new tab
			setAttribute(n, "target", "_blank")
			setAttribute(n, "rel", "noopener")
		}
	}

	// Recurse
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		updateArticleDom(c)
	}
}

func FindArticleImage(body string) string {

	body = strings.ReplaceAll(body, "{STEAM_CLAN_IMAGE}", articleImageBase)
	body = strings.ReplaceAll(body, "{STEAM_CLAN_LOC_IMAGE}", articleImageBase)

	body = BBCodeCompiler.Compile(body)

	doc, err := html.Parse(strings.NewReader(body))
	if err == nil {
		src := findArticleImage(doc)
		if src != "" {
			return src
		}
	}

	return ""
}

func findArticleImage(n *html.Node) string {

	// Recurse
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		src := findArticleImage(c)
		if src != "" {
			return src
		}
	}

	// Get image src
	if n.Type == html.ElementNode && n.Data == "img" {

		for _, attr := range n.Attr {
			// Find image that images.weserv.nl supports
			if attr.Key == "src" &&
				(strings.HasSuffix(attr.Val, "png") ||
					strings.HasSuffix(attr.Val, "jpg") ||
					strings.HasSuffix(attr.Val, "jpeg") ||
					strings.HasSuffix(attr.Val, "bmp") ||
					strings.HasSuffix(attr.Val, "webp") ||
					strings.HasSuffix(attr.Val, "svg")) {
				return attr.Val
			}
		}
	}

	return ""
}
