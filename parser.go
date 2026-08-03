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
	htmlTagPattern      = regexp.MustCompile(`<[^>]+>`)
	templateCallPattern = regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)
)

func newExtractorConfig() extractorConfig {
	return extractorConfig{
		keywords:   []string{"t", "T", "gettext", "i18n"},
		exts:       []string{".go", ".gohtml", ".html"},
		ignoreDirs: []string{"vendor", ".git", "node_modules"},
	}
}

func extractGoSource(src string) []string {
	return extractMessageIDs(extractGoSourceMessagesWithConfig(src, newExtractorConfig()))
}

func extractGoSourceWithConfig(src string, cfg extractorConfig) []string {
	return extractMessageIDs(extractGoSourceMessagesWithConfig(src, cfg))
}

func extractGoSourceMessagesWithConfig(src string, cfg extractorConfig) []message {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil
	}

	var messages []message
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
		line := fset.Position(call.Pos()).Line
		messages = append(messages, message{id: value, line: line})
		return true
	})

	return messages
}

func extractMessageIDs(messages []message) []string {
	if messages == nil {
		return nil
	}
	ids := make([]string, 0, len(messages))
	for _, msg := range messages {
		ids = append(ids, msg.id)
	}
	return ids
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
			return containsKeyword(keywords, ident.Name+"."+fn.Sel.Name)
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
	return extractMessageIDs(extractGoHTMLTemplateMessagesWithConfig(src, newExtractorConfig()))
}

func extractGoHTMLTemplateMessagesWithConfig(src string, cfg extractorConfig) []message {
	var messages []message
	seen := make(map[string]struct{})
	add := func(value string, line int) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, already := seen[value]; already {
			return
		}
		seen[value] = struct{}{}
		messages = append(messages, message{id: value, line: line})
	}

	actionMatches := templateCallPattern.FindAllStringSubmatchIndex(src, -1)
	for _, action := range actionMatches {
		if len(action) < 4 {
			continue
		}
		actionBodyStart := action[2]
		actionBodyEnd := action[3]
		actionBody := src[actionBodyStart:actionBodyEnd]
		for _, keyword := range cfg.keywords {
			pattern := regexp.MustCompile(`(?:^|[^\w.])(?:\.)?` + regexp.QuoteMeta(keyword) + `\s+"((?:[^"\\]|\\.)*)"`)
			for _, tpl := range pattern.FindAllStringSubmatchIndex(actionBody, -1) {
				if len(tpl) < 4 {
					continue
				}
				messageStart := actionBodyStart + tpl[2]
				line := 1 + strings.Count(src[:messageStart], "\n")
				add(actionBody[tpl[2]:tpl[3]], line)
			}
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
