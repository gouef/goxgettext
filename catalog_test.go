package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildPOTIncludesReferences(t *testing.T) {
	entries := []message{{id: "Hello", file: "main.go", line: 7}}

	got := buildPOT(entries)

	if !strings.Contains(got, "#: main.go:7") {
		t.Fatalf("buildPOT() missing reference, got %q", got)
	}
	if !strings.Contains(got, `msgid "Hello"`) {
		t.Fatalf("buildPOT() missing message id, got %q", got)
	}
}

func TestBuildPOIncludesHeader(t *testing.T) {
	got := buildPO([]message{{id: "Hello"}})

	if !strings.Contains(got, `msgid ""`) {
		t.Fatalf("buildPO() missing header, got %q", got)
	}
}

func TestCollectMessagesAcrossMixedFiles(t *testing.T) {
	tempDir := t.TempDir()

	goSrc := `package main

func main() {
	_ = gettext("Hello")
}
`
	goHTMLSrc := `<h1>Hello</h1>
<p>Welcome</p>
{{ if .User }}<span>{{ t "Bonjour" }}</span>{{ end }}
`

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goSrc), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "page.gohtml"), []byte(goHTMLSrc), 0o644); err != nil {
		t.Fatalf("write page.gohtml: %v", err)
	}

	entries, err := collectMessages(tempDir)
	if err != nil {
		t.Fatalf("collectMessages() error = %v", err)
	}

	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.id)
	}

	want := []string{"Bonjour", "Hello"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectMessages() = %v, want %v", got, want)
	}
}

func TestCollectMessagesRecursesIntoSubdirectories(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Nested")
}
`), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	entries, err := collectMessages(tempDir)
	if err != nil {
		t.Fatalf("collectMessages() error = %v", err)
	}

	if len(entries) != 1 || entries[0].id != "Nested" {
		t.Fatalf("collectMessages() = %v, want one nested message", entries)
	}
}

func TestCollectMessagesReturnsErrorForMissingPath(t *testing.T) {
	_, err := collectMessages(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("collectMessages() error = nil, want an error for missing path")
	}
}

func TestMessageSetAddHandlesExistingEntryAndEmptyIDs(t *testing.T) {
	set := newMessageSet()
	set.add("Hello", "", 0)
	set.add("Hello", "main.go", 2)
	set.add("", "ignored.go", 3)

	if len(set.entries) != 1 {
		t.Fatalf("messageSet length = %d, want 1", len(set.entries))
	}

	entry := set.entries["Hello"]
	if entry.file != "main.go" || entry.line != 2 {
		t.Fatalf("messageSet entry = %+v, want file main.go and line 2", entry)
	}
}

func TestBuildPOTAndBuildPOSkipEmptyMessages(t *testing.T) {
	entries := []message{{id: "", file: "main.go", line: 1}, {id: "Hello", file: "main.go", line: 2}}

	pot := buildPOT(entries)
	if strings.Contains(pot, "#: main.go:1") || strings.Contains(pot, `msgid ""`) {
		t.Fatalf("buildPOT() should skip empty ids, got %q", pot)
	}
	if !strings.Contains(pot, `msgid "Hello"`) {
		t.Fatalf("buildPOT() missing message id, got %q", pot)
	}

	po := buildPO(entries)
	if !strings.Contains(po, `msgid "Hello"`) {
		t.Fatalf("buildPO() missing message id, got %q", po)
	}
}

func TestCollectMessagesIncludesLineNumbers(t *testing.T) {
	tempDir := t.TempDir()

	goSrc := `package main

func main() {
	_ = gettext("GoLine")
}
`
	tplSrc := `<div>
{{ t "TplLine" }}
</div>`

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goSrc), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "page.gohtml"), []byte(tplSrc), 0o644); err != nil {
		t.Fatalf("write page.gohtml: %v", err)
	}

	entries, err := collectMessages(tempDir)
	if err != nil {
		t.Fatalf("collectMessages() error = %v", err)
	}

	lines := make(map[string]int)
	for _, entry := range entries {
		lines[entry.id] = entry.line
	}

	if lines["GoLine"] != 4 {
		t.Fatalf("GoLine line = %d, want 4", lines["GoLine"])
	}
	if lines["TplLine"] != 2 {
		t.Fatalf("TplLine line = %d, want 2", lines["TplLine"])
	}
}
