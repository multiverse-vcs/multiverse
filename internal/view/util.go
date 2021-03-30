package view

import (
	"html/template"
	"path"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

const syntaxStyle = "monokai"

var formatter = html.New(
	html.WithLineNumbers(true),
	html.LineNumbersInTable(true),
	html.LinkableLineNumbers(true, ""),
)

var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithStyle(syntaxStyle),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

// Highlight renders the given source into syntax highlighted HTML.
func Highlight(name string, source []byte) (template.HTML, error) {
	lexer := lexers.Match(name)
	if lexer == nil {
		lexer = lexers.Analyse(string(source))
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}

	lexer = chroma.Coalesce(lexer)
	style := styles.Get(syntaxStyle)

	iterator, err := lexer.Tokenise(nil, string(source))
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := formatter.Format(&result, style, iterator); err != nil {
		return "", err
	}

	return template.HTML(result.String()), nil
}

// Markdown renders the given markdown source into HTML.
func Markdown(source []byte) (template.HTML, error) {
	var result strings.Builder
	if err := markdown.Convert(source, &result); err != nil {
		return "", err
	}

	return template.HTML(result.String()), nil
}

// Breadcrumbs returns a list of breadcrumbs for the given URL.
func Breadcrumbs(url string) []string {
	var breadcrumbs []string
	for i := url; i != "/" && i != "."; i = path.Dir(i) {
		breadcrumbs = append(breadcrumbs, i)
	}

	return breadcrumbs
}
