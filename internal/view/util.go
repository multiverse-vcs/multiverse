package view

import (
	"html/template"
	"io"
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

var goldmarkdown = goldmark.New(
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

// highlight renders the given reader into highlighted HTML.
func highlight(name string, r io.Reader) (template.HTML, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

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

// markdown renders the given reader into HTML.
func markdown(r io.Reader) (template.HTML, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := goldmarkdown.Convert(source, &result); err != nil {
		return "", err
	}

	return template.HTML(result.String()), nil
}

// breadcrumbs returns a list of breadcrumbs for the given URL.
func breadcrumbs(url string) []string {
	var breadcrumbs []string
	for i := url; i != "/" && i != "."; i = path.Dir(i) {
		breadcrumbs = append(breadcrumbs, i)
	}

	return breadcrumbs
}
