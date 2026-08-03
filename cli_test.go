package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLIWritesPOTFile(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.pot")

	src := `package main

func main() {
	_ = gettext("Hello")
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	if err := runCLI([]string{"--output", outputPath, tempDir}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("output file missing message, got %q", string(content))
	}
}

func TestRunCLIWithPOFormat(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.po")

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	if err := runCLI([]string{"--format", "po", "--output", outputPath, tempDir}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}

	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("PO output missing message, got %q", string(content))
	}
}

func TestRunCLIWritesLanguageFilesAndPreservesTranslations(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.pot")
	localeDir := filepath.Join(tempDir, "locale")
	languageFile := filepath.Join(localeDir, "cs.po")

	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("mkdir locale dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	if err := os.WriteFile(languageFile, []byte(`msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\\n"
"Language: cs\\n"

msgid "Hello"
msgstr "Ahoj"
`), 0o644); err != nil {
		t.Fatalf("write initial language file: %v", err)
	}

	if err := runCLI([]string{"--output", outputPath, "--output-dir", localeDir, "--language", "cs", tempDir}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	content, err := os.ReadFile(languageFile)
	if err != nil {
		t.Fatalf("read language file: %v", err)
	}

	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("language file missing msgid, got %q", string(content))
	}
	if !strings.Contains(string(content), `msgstr "Ahoj"`) {
		t.Fatalf("language file did not preserve translation, got %q", string(content))
	}
}

func TestRunCLIGeneratesLanguagesFromLINGUASAndPOTFILES(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "locale")
	linguasPath := filepath.Join(tempDir, "LINGUAS")

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(`<html><body>Welcome</body></html>`), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(linguasPath, []byte("cs\nfr\n"), 0o644); err != nil {
		t.Fatalf("write LINGUAS: %v", err)
	}

	if err := runCLI([]string{"--output-dir", outputDir, tempDir}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	for _, language := range []string{"cs", "fr"} {
		path := filepath.Join(outputDir, language+".po")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected language file %s to exist: %v", path, err)
		}
	}

	potfilesPath := filepath.Join(outputDir, "POTFILES")
	content, err := os.ReadFile(potfilesPath)
	if err != nil {
		t.Fatalf("read POTFILES: %v", err)
	}
	if !strings.Contains(string(content), "main.go") || !strings.Contains(string(content), "index.html") {
		t.Fatalf("POTFILES missing expected entries, got %q", string(content))
	}
}

func TestRunCLIRerunAfterMessageMovesOrDisappears(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.pot")
	mainPath := filepath.Join(tempDir, "main.go")

	initialSrc := `package main

func main() {
	_ = gettext("Hello")
}
`
	if err := os.WriteFile(mainPath, []byte(initialSrc), 0o644); err != nil {
		t.Fatalf("write initial main.go: %v", err)
	}

	if err := runCLI([]string{"--output", outputPath, tempDir}); err != nil {
		t.Fatalf("first runCLI() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output after first run: %v", err)
	}
	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("first output missing message, got %q", string(content))
	}

	shiftedSrc := `package main

func main() {
	fmt.Println("moved")
	_ = gettext("Hello")
}
`
	if err := os.WriteFile(mainPath, []byte(shiftedSrc), 0o644); err != nil {
		t.Fatalf("write shifted main.go: %v", err)
	}

	if err := runCLI([]string{"--output", outputPath, tempDir}); err != nil {
		t.Fatalf("second runCLI() error = %v", err)
	}

	content, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output after second run: %v", err)
	}
	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("second output missing message after shift, got %q", string(content))
	}

	removedSrc := `package main

func main() {
	fmt.Println("removed")
}
`
	if err := os.WriteFile(mainPath, []byte(removedSrc), 0o644); err != nil {
		t.Fatalf("write removed main.go: %v", err)
	}

	if err := runCLI([]string{"--output", outputPath, tempDir}); err != nil {
		t.Fatalf("third runCLI() error = %v", err)
	}

	content, err = os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output after third run: %v", err)
	}
	if strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("third output should not contain removed message, got %q", string(content))
	}
}

func TestMainRunsCLIInSubprocess(t *testing.T) {
	if os.Getenv("GOXGETTEXT_HELPER_PROCESS") == "1" {
		outputPath := os.Getenv("GOXGETTEXT_OUTPUT")
		if outputPath == "" {
			fmt.Fprintln(os.Stderr, "missing output path")
			os.Exit(2)
		}
		inputDir := os.Getenv("GOXGETTEXT_INPUT_DIR")
		if inputDir == "" {
			fmt.Fprintln(os.Stderr, "missing input dir")
			os.Exit(2)
		}
		if err := os.WriteFile(filepath.Join(inputDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Args = []string{"goxgettext", "--output", outputPath, inputDir}
		main()
		return
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.pot")
	cmd := exec.Command(os.Args[0], "-test.run=TestMainRunsCLIInSubprocess$")
	cmd.Env = append(os.Environ(), "GOXGETTEXT_HELPER_PROCESS=1", "GOXGETTEXT_INPUT_DIR="+tempDir, "GOXGETTEXT_OUTPUT="+outputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper process failed: %v\n%s", err, out)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("output file missing message, got %q", string(content))
	}
}

func TestMainExitsOnCLIErrorInSubprocess(t *testing.T) {
	if os.Getenv("GOXGETTEXT_HELPER_PROCESS_ERROR") == "1" {
		os.Args = []string{"goxgettext", "--output", filepath.Join(os.Getenv("GOXGETTEXT_INPUT_DIR"), "messages.pot"), "does-not-exist"}
		main()
		return
	}

	tempDir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitsOnCLIErrorInSubprocess$")
	cmd.Env = append(os.Environ(), "GOXGETTEXT_HELPER_PROCESS_ERROR=1", "GOXGETTEXT_INPUT_DIR="+tempDir)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected helper process to fail")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v", err)
	}
}

func TestRunCLIWithAllFlagGeneratesPOTPOAndLanguages(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "locale")

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "LINGUAS"), []byte("cs\n"), 0o644); err != nil {
		t.Fatalf("write LINGUAS: %v", err)
	}

	if err := runCLI([]string{"--all", "--output-dir", outputDir, tempDir}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "messages.pot")); err != nil {
		t.Fatalf("expected POT file to be created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "cs.po")); err != nil {
		t.Fatalf("expected language file to be created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "POTFILES")); err != nil {
		t.Fatalf("expected POTFILES to be created: %v", err)
	}
}

func TestRunCLIUsesCurrentDirectoryWhenNoPathsProvided(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "messages.pot")
	mainPath := filepath.Join(tempDir, "main.go")

	if err := os.WriteFile(mainPath, []byte(`package main

func main() {
	_ = gettext("Hello")
}
`), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if err := runCLI([]string{"--output", outputPath}); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(content), `msgid "Hello"`) {
		t.Fatalf("output file missing message, got %q", string(content))
	}
}

func TestRunCLIReportsMissingPath(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "messages.pot")

	if err := runCLI([]string{"--output", outputPath, "does-not-exist"}); err == nil {
		t.Fatal("expected runCLI() to fail for a missing path")
	}
}
