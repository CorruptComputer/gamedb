package helpers

import (
	"html"
	"html/template"

	"github.com/frustra/bbcode"
)

var BBCodeCompiler = bbcode.NewCompiler(true, true)

func init() {

	BBCodeCompiler.SetTag("h1", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "h1"
		return out, true
	})
	BBCodeCompiler.SetTag("strike", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "span"
		out.Attrs["style"] = "text-decoration: line-through;"
		return out, true
	})
	BBCodeCompiler.SetTag("list", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "ul"
		return out, true
	})
	BBCodeCompiler.SetTag("*", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "li"
		return out, true
	})
	BBCodeCompiler.SetTag("table", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "table"
		out.Attrs["class"] = "table table-hover table-striped mb-0"
		return out, true
	})
	BBCodeCompiler.SetTag("tr", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "tr"
		return out, true
	})
	BBCodeCompiler.SetTag("td", func(node *bbcode.BBCodeNode) (*bbcode.HTMLTag, bool) {
		out := bbcode.NewHTMLTag("")
		out.Name = "td"
		return out, true
	})
}

func RenderHTMLAndBBCode(in string) template.HTML {

	return template.HTML(html.UnescapeString(BBCodeCompiler.Compile(in)))

}
