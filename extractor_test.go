package main

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestExtractGoSource(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	_ = gettext("Hello")
	_ = T("World")
	_ = fmt.Sprintf("skip me")
}
`

	got := extractGoSource(src)
	want := []string{"Hello", "World"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoSource() = %v, want %v", got, want)
	}
}

func TestExtractGoHTMLTemplate(t *testing.T) {
	src := `<h1>Hello</h1>
<p>Welcome</p>
{{ if .User }}<span>{{ t "Bonjour" }}</span>{{ end }}
`

	got := extractGoHTMLTemplate(src)
	want := []string{"Hello", "Welcome", "Bonjour"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoHTMLTemplate() = %v, want %v", got, want)
	}
}

func TestExtractGoSourceSupportsSelectorCalls(t *testing.T) {
	src := `package main

func main() {
	_ = t("Hello")
	_ = i18n.T("World")
	_ = gettext("Again")
}
`

	got := extractGoSource(src)
	want := []string{"Hello", "World", "Again"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoSource() = %v, want %v", got, want)
	}
}

func TestExtractGoSourceWithCustomKeyword(t *testing.T) {
	src := `package main

func main() {
	_ = hello("Custom")
}
`

	got := extractGoSourceWithConfig(src, extractorConfig{keywords: []string{"hello"}})
	want := []string{"Custom"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoSourceWithConfig() = %v, want %v", got, want)
	}
}

func TestExtractGoSourceIncludesI18nKeyword(t *testing.T) {
	src := `package main

func main() {
	_ = i18n.T("Hello")
}
`

	got := extractGoSource(src)
	want := []string{"Hello"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoSource() = %v, want %v", got, want)
	}
}

func TestExtractHTMLTemplateText(t *testing.T) {
	src := `<html><body><h1>Welcome</h1><p>Home page</p></body></html>`

	got := extractGoHTMLTemplate(src)
	want := []string{"Welcome", "Home page"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoHTMLTemplate() = %v, want %v", got, want)
	}
}

func TestIsTranslationCallSupportsIdentifiersAndSelectors(t *testing.T) {
	if !isTranslationCall(ast.NewIdent("gettext"), []string{"gettext"}) {
		t.Fatal("isTranslationCall() should accept plain identifiers")
	}

	selector := &ast.SelectorExpr{X: ast.NewIdent("i18n"), Sel: ast.NewIdent("T")}
	if !isTranslationCall(selector, []string{"T"}) {
		t.Fatal("isTranslationCall() should accept selector expressions")
	}

	if isTranslationCall(selector, []string{"gettext"}) {
		t.Fatal("isTranslationCall() should reject unsupported selector names")
	}
}

func TestStringLiteralValueRejectsNonStringLiterals(t *testing.T) {
	if got, ok := stringLiteralValue(&ast.BasicLit{Kind: token.INT, Value: "123"}); ok || got != "" {
		t.Fatalf("stringLiteralValue() = %q, %v, want empty and false", got, ok)
	}

	literal := &ast.BasicLit{Kind: token.STRING, Value: `"Hello"`}
	got, ok := stringLiteralValue(literal)
	if !ok || got != "Hello" {
		t.Fatalf("stringLiteralValue() = %q, %v, want %q, true", got, ok, "Hello")
	}
}

func TestExtractVisibleTextHandlesTemplateControl(t *testing.T) {
	got := extractVisibleText(`<div>{{ if .User }}<span>Visible</span>{{ else }}<span>Other</span>{{ end }}</div>`)
	want := []string{"Visible", "Other"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractVisibleText() = %v, want %v", got, want)
	}
}

func TestExtractGoSourceIgnoresMalformedCalls(t *testing.T) {
	src := `package main

func main() {
	_ = gettext()
	_ = gettext(123)
	_ = gettext("Hello")
}
`

	got := extractGoSource(src)
	want := []string{"Hello"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractGoSource() = %v, want %v", got, want)
	}
}

func TestExtractGoSourceWithConfigReturnsNilForParseErrors(t *testing.T) {
	got := extractGoSourceWithConfig("package main\nfunc main(", extractorConfig{})
	if got != nil {
		t.Fatalf("extractGoSourceWithConfig() = %v, want nil", got)
	}
}

func TestExtractVisibleTextReturnsNilForEmptyInput(t *testing.T) {
	if got := extractVisibleText(""); got != nil {
		t.Fatalf("extractVisibleText() = %v, want nil", got)
	}
}
