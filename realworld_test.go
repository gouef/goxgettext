package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealWorldTemplateAndHTMLScanning(t *testing.T) {
	tempDir := t.TempDir()

	goFile := `package main

import "fmt"

func main() {
	_ = gettext("Welcome")
	_ = i18n.T("Goodbye")
	fmt.Println("ignored")
}
`
	htmlFile := `<html>
  <body>
    <h1>Hello</h1>
    <p>World</p>
  </body>
</html>`

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goFile), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(htmlFile), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	entries, err := collectMessagesWithConfig(tempDir, extractorConfig{keywords: []string{"gettext", "i18n"}, exts: []string{".go", ".html"}})
	if err != nil {
		t.Fatalf("collectMessagesWithConfig() error = %v", err)
	}

	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.id)
	}

	for _, want := range []string{"Welcome", "Goodbye", "Hello", "World"} {
		if !contains(ids, want) {
			t.Fatalf("missing %q in %v", want, ids)
		}
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestRunCLIWithExtensionFlag(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.pot")

	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(`<html><body>Hello</body></html>`), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	if err := runCLI([]string{"--extension", ".html", "--output", outputPath, tempDir}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("extension output missing message, got %q", string(content))
	}
}
