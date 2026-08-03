package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveLanguagesReadsLINGUAS(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "LINGUAS"), []byte("cs\nen\n"), 0o644); err != nil {
		t.Fatalf("write LINGUAS: %v", err)
	}

	got, err := resolveLanguages([]string{tempDir})
	if err != nil {
		t.Fatalf("resolveLanguages() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{"cs", "en"}) {
		t.Fatalf("resolveLanguages() = %v, want %v", got, []string{"cs", "en"})
	}
}

func TestReadLINGUASSkipsCommentsAndBlankLines(t *testing.T) {
	tempDir := t.TempDir()
	linguasPath := filepath.Join(tempDir, "LINGUAS")
	if err := os.WriteFile(linguasPath, []byte("# comment\ncs\n\nfr\n"), 0o644); err != nil {
		t.Fatalf("write LINGUAS: %v", err)
	}

	got, err := readLINGUAS([]string{tempDir})
	if err != nil {
		t.Fatalf("readLINGUAS() error = %v", err)
	}
	want := []string{"cs", "fr"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readLINGUAS() = %v, want %v", got, want)
	}
}

func TestWriteLanguageFilePreservesExistingTranslations(t *testing.T) {
	tempDir := t.TempDir()
	languagePath := filepath.Join(tempDir, "cs.po")
	if err := os.WriteFile(languagePath, []byte(`msgid "Hello"
msgstr "Ahoj"
`), 0o644); err != nil {
		t.Fatalf("write existing language file: %v", err)
	}

	entries := []message{{id: "Hello"}, {id: "World"}}
	if err := writeLanguageFile(tempDir, "cs", entries); err != nil {
		t.Fatalf("writeLanguageFile() error = %v", err)
	}

	content, err := os.ReadFile(languagePath)
	if err != nil {
		t.Fatalf("read updated language file: %v", err)
	}
	if !strings.Contains(string(content), `msgstr "Ahoj"`) {
		t.Fatalf("expected preserved translation, got %q", string(content))
	}
	if !strings.Contains(string(content), `msgid "World"`) {
		t.Fatalf("expected new message entry, got %q", string(content))
	}
}

func TestWritePOTFILESCollectsSupportedFiles(t *testing.T) {
	tempDir := t.TempDir()
	outDir := filepath.Join(tempDir, "locale")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "page.gohtml"), []byte(`{{ t "Hi" }}`), 0o644); err != nil {
		t.Fatalf("write page.gohtml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "ignored.gohtml"), []byte(`<h1>plain text</h1>`), 0o644); err != nil {
		t.Fatalf("write ignored.gohtml: %v", err)
	}

	if err := writePOTFILES(outDir, []string{tempDir}, newExtractorConfig()); err != nil {
		t.Fatalf("writePOTFILES() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "POTFILES"))
	if err != nil {
		t.Fatalf("read POTFILES: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "main.go") || !strings.Contains(text, "page.gohtml") {
		t.Fatalf("POTFILES missing expected entries, got %q", text)
	}
	if strings.Contains(text, "notes.txt") {
		t.Fatalf("POTFILES should not include unsupported file, got %q", text)
	}
	if strings.Contains(text, "ignored.gohtml") {
		t.Fatalf("POTFILES should not include files without extractable messages, got %q", text)
	}
}

func TestCollectSourceFilesReturnsFilesForDirectoryAndSingleFile(t *testing.T) {
	tempDir := t.TempDir()
	mainPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	files, err := collectSourceFiles([]string{mainPath}, newExtractorConfig())
	if err != nil {
		t.Fatalf("collectSourceFiles(single file) error = %v", err)
	}
	if !reflect.DeepEqual(files, []string{mainPath}) {
		t.Fatalf("collectSourceFiles(single file) = %v, want %v", files, []string{mainPath})
	}

	dirFiles, err := collectSourceFiles([]string{tempDir}, newExtractorConfig())
	if err != nil {
		t.Fatalf("collectSourceFiles(directory) error = %v", err)
	}
	if len(dirFiles) != 1 || dirFiles[0] != mainPath {
		t.Fatalf("collectSourceFiles(directory) = %v, want %v", dirFiles, []string{mainPath})
	}
}

func TestParseExistingTranslationsHandlesMultipleEntries(t *testing.T) {
	content := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"

msgid "Hello"
msgstr "Ahoj"

msgid "World"
msgstr "Svět"
`

	got := parseExistingTranslations(content)
	want := map[string]string{"Hello": "Ahoj", "World": "Svět"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseExistingTranslations() = %v, want %v", got, want)
	}
}

func TestParsePOValueHandlesQuotedAndUnquoted(t *testing.T) {
	if got := parsePOValue(`"Hello"`); got != "Hello" {
		t.Fatalf("parsePOValue(quoted) = %q, want %q", got, "Hello")
	}
	if got := parsePOValue("Hello"); got != "Hello" {
		t.Fatalf("parsePOValue(unquoted) = %q, want %q", got, "Hello")
	}
}
