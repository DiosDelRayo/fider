package markdown

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"io"
	"regexp"
	"strings"
)

func textRenderLink(w io.Writer, link *ast.Link, entering bool) {
	if !entering {
		return
	}
	w.Write([]byte(fmt.Sprintf(`%s: %s`, string(link.Literal), link.Destination)))
}

func textRenderNodeHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch node := node.(type) {
	case *ast.HTMLSpan:
		html.EscapeHTML(w, node.Literal)
		return ast.GoToNext, true
	case *ast.HTMLBlock:
		_, _ = io.WriteString(w, "\n")
		html.EscapeHTML(w, node.Literal)
		_, _ = io.WriteString(w, "\n")
		return ast.GoToNext, true
	case *ast.Code:
		_, _ = io.WriteString(w, fmt.Sprintf("`%s`", node.Literal))
		return ast.GoToNext, true
	case *ast.Link:
		uri := node.Destination
		if !(moneroLinkRegex.MatchString(string(uri)) || bitcoinLinkRegex.MatchString(string(uri))) {
			// Let the default link renderer handle the link
			return ast.GoToNext, false
		}
		textRenderLink(w, node, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

var textRenderer = html.NewRenderer(html.RendererOptions{
	Flags:	html.UseXHTML |
		html.Smartypants |
		html.SmartypantsFractions |
		html.SmartypantsDashes |
		html.SmartypantsLatexDashes |
		html.Safelink |
		html.NofollowLinks |
		html.NoreferrerLinks,
	RenderNodeHook: textRenderNodeHook,
})

// The policy strips all HTML tags from the input text.
var strictPolicy = bluemonday.StrictPolicy()

// The regular expression finds duplicate newlines.
var regexNewlines = regexp.MustCompile(`\n+`)

// PlainText parses given markdown input and return only the text
func PlainText(input string) string {
	// Apparently a parser cannot be reused.
	// https://github.com/gomarkdown/markdown/issues/229
	output := markdown.ToHTML([]byte(input), parser.NewWithExtensions(parser.CommonExtensions | parser.Autolink), textRenderer)
	sanitizedOutput := strictPolicy.Sanitize(string(output))
	sanitizedOutput = regexNewlines.ReplaceAllString(sanitizedOutput, "\n")
	return strings.TrimSpace(sanitizedOutput)
}
