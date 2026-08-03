package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

type extractorConfig struct {
	keywords   []string
	exts       []string
	ignoreDirs []string
}

var (
	htmlTagPattern        = regexp.MustCompile(`<[^>]+>`)
	templateCallPattern   = regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)
	templateStringPattern = regexp.MustCompile(`(?:^|[^\w.])(?:t|T|gettext|\.t|\.T|\.gettext)\s+"((?:[^"\\]|\\.)*)"`)
)

func newExtractorConfig() extractorConfig {
	return extractorConfig{
		keywords:   []string{"t", "T", "gettext", "i18n"},
		exts:       []string{".go", ".gohtml", ".html"},
		ignoreDirs: []string{"vendor"},
	}
}

func extractGoSource(src string) []string {
	return extractGoSourceWithConfig(src, newExtractorConfig())
}

func extractGoSourceWithConfig(src string, cfg extractorConfig) []string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil
	}

	var messages []string
	seen := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !isTranslationCall(call.Fun, cfg.keywords) {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		value, ok := stringLiteralValue(call.Args[0])
		if !ok {
			return true
		}
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
		messages = append(messages, value)
		return true
	})

	return messages
}

func isTranslationCall(expr ast.Expr, keywords []string) bool {
	switch fn := expr.(type) {
	case *ast.Ident:
		return containsKeyword(keywords, fn.Name)
	case *ast.SelectorExpr:
		if containsKeyword(keywords, fn.Sel.Name) {
			return true
		}
		if ident, ok := fn.X.(*ast.Ident); ok {
			return containsKeyword(keywords, ident.Name)
		}
		return false
	default:
		return false
	}
}

func containsKeyword(keywords []string, name string) bool {
	for _, keyword := range keywords {
		if keyword == name {
			return true
		}
	}
	return false
}

func stringLiteralValue(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value := lit.Value
	if len(value) < 2 {
		return "", false
	}
	unquoted, err := strconv.Unquote(value)
	if err != nil {
		return "", false
	}
	return unquoted, true
}

func extractGoHTMLTemplate(src string) []string {
	var messages []string
	seen := make(map[string]struct{})
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, already := seen[value]; already {
			return
		}
		seen[value] = struct{}{}
		messages = append(messages, value)
	}

	for _, part := range extractVisibleText(src) {
		add(part)
	}

	for _, match := range templateCallPattern.FindAllStringSubmatch(src, -1) {
		if len(match) < 2 {
			continue
		}
		for _, tpl := range templateStringPattern.FindAllStringSubmatch(match[1], -1) {
			if len(tpl) < 2 {
				continue
			}
			add(tpl[1])
		}
	}

	return messages
}

func extractVisibleText(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	withoutDelims := strings.ReplaceAll(text, "{{", " ")
	withoutDelims = strings.ReplaceAll(withoutDelims, "}}", " ")
	withoutDelims = strings.ReplaceAll(withoutDelims, "\r\n", "\n")
	withoutDelims = strings.ReplaceAll(withoutDelims, "\r", "\n")
	parts := strings.Split(htmlTagPattern.ReplaceAllString(withoutDelims, "\n"), "\n")
	var out []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if strings.ContainsAny(trimmed, "<>/{}()\"'") || strings.HasPrefix(trimmed, "if") || strings.HasPrefix(trimmed, "end") || strings.HasPrefix(trimmed, "else") || strings.HasPrefix(trimmed, "range") || strings.HasPrefix(trimmed, "with") || strings.HasPrefix(trimmed, "template") || strings.HasPrefix(trimmed, ".") || trimmed == "t" || trimmed == "T" || trimmed == "gettext" {
			continue
		}
		out = append(out, strings.Join(strings.Fields(trimmed), " "))
	}
	return out
}
