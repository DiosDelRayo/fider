package markdown

import (
	"html/template"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var (
	moneroLinkRegex = regexp.MustCompile(`^monero:(\/\/)?([a-zA-Z0-9]{95})(\?(tx_amount|recipient_name|tx_description)=[^&]+((&tx_amount|&recipient_name|&tx_description)=[^&]+)*)*$`)
	bitcoinLinkRegex = regexp.MustCompile(`^bitcoin:([a-zA-Z0-9]{26,35}|[a-zA-Z0-9]{42})(\?(amount|label|message)=[^&]+((&amount|&label|&message)=[^&]+)*)*$`)
)

func renderLink(w io.Writer, link *ast.Link, entering bool) {
	if !entering {
		w.Write([]byte(`</a>`))
		return
	}
	w.Write([]byte(fmt.Sprintf(`<a href="%s" rel="noopener">%s`, link.Destination, string(link.Literal))))
}

func renderNodeHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch node := node.(type) {
	case *ast.HTMLSpan:
		html.EscapeHTML(w, node.Literal)
		return ast.GoToNext, true
	case *ast.HTMLBlock:
		w.Write([]byte("\n"))
		html.EscapeHTML(w, node.Literal)
		w.Write([]byte("\n"))
		return ast.GoToNext, true
	case *ast.Link:
		uri := node.Destination
		if !(moneroLinkRegex.MatchString(string(uri)) || bitcoinLinkRegex.MatchString(string(uri))) {
			// Let the default link renderer handle the link
			return ast.GoToNext, false
		}
		renderLink(w, node, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

// Full turns a markdown into HTML using all rules
func Full(input string) template.HTML {
	parser := parser.NewWithExtensions(parser.CommonExtensions | parser.Autolink)
	renderer := html.NewRenderer(html.RendererOptions{
		RenderNodeHook: renderNodeHook,
		Flags:	html.UseXHTML |
			html.Smartypants |
			html.SmartypantsFractions |
			html.SmartypantsDashes |
			html.SmartypantsLatexDashes |
			html.Safelink |
			html.NofollowLinks |
			html.NoreferrerLinks,
	})
	return template.HTML(strings.TrimSpace(string(markdown.ToHTML([]byte(input), parser, renderer))))
}
