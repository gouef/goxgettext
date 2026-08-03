package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func resolveLanguages(paths []string, outputDir string) ([]string, error) {
	if outputDir != "" {
		languages, err := readLINGUAS([]string{outputDir})
		if err != nil {
			return nil, err
		}
		if len(languages) > 0 {
			return languages, nil
		}
	}
	return readLINGUAS(paths)
}

func writeLanguageFiles(outputDir string, entries []message, languages []string) error {
	for _, language := range languages {
		if err := writeLanguageFile(outputDir, language, entries); err != nil {
			return err
		}
	}
	return nil
}

func writeLanguageFile(outputDir, language string, entries []message) error {
	languagePath := filepath.Join(outputDir, language+".po")
	translations := map[string]string{}
	if existing, err := os.ReadFile(languagePath); err == nil && len(existing) > 0 {
		translations = parseExistingTranslations(string(existing))
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := buildLanguagePO(entries, translations, language)
	return os.WriteFile(languagePath, []byte(content), 0o644)
}

func readLINGUAS(paths []string) ([]string, error) {
	for _, path := range paths {
		linguasPath := filepath.Join(path, "LINGUAS")
		file, err := os.Open(linguasPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		defer file.Close()

		var languages []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			languages = append(languages, line)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		if len(languages) > 0 {
			return languages, nil
		}
	}
	return nil, nil
}

func writePOTFILES(outputDir string, paths []string, cfg extractorConfig) error {
	files, err := collectSourceFiles(paths, cfg)
	if err != nil {
		return err
	}

	potfilesPath := filepath.Join(outputDir, "POTFILES")
	content := strings.Join(files, "\n") + "\n"
	return os.WriteFile(potfilesPath, []byte(content), 0o644)
}

func collectSourceFiles(paths []string, cfg extractorConfig) ([]string, error) {
	var files []string
	collectIfTranslatable := func(current string) error {
		ext := strings.ToLower(filepath.Ext(current))
		if !hasExtension(cfg.exts, ext) {
			return nil
		}

		content, err := os.ReadFile(current)
		if err != nil {
			return err
		}

		switch ext {
		case ".go":
			if len(extractGoSourceMessagesWithConfig(string(content), cfg)) > 0 {
				files = append(files, current)
			}
		case ".gohtml", ".html":
			if len(extractGoHTMLTemplateMessagesWithConfig(string(content), cfg)) > 0 {
				files = append(files, current)
			}
		}
		return nil
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if err := collectIfTranslatable(path); err != nil {
				return nil, err
			}
			continue
		}
		walkErr := filepath.Walk(path, func(current string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				for _, ignored := range newExtractorConfig().ignoreDirs {
					if strings.EqualFold(info.Name(), ignored) {
						return filepath.SkipDir
					}
				}
				return nil
			}
			return collectIfTranslatable(current)
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}
	return files, nil
}

func buildLanguagePO(entries []message, translations map[string]string, language string) string {
	var b strings.Builder
	b.WriteString("msgid \"\"\n")
	b.WriteString("msgstr \"\"\n")
	fmt.Fprintf(&b, "\"Content-Type: text/plain; charset=UTF-8\\n\"\n")
	fmt.Fprintf(&b, "\"Language: %s\\n\"\n\n", language)
	for _, entry := range entries {
		if entry.id == "" {
			continue
		}
		msgstr := translations[entry.id]
		fmt.Fprintf(&b, "msgid %q\nmsgstr %q\n\n", entry.id, msgstr)
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func parseExistingTranslations(content string) map[string]string {
	translations := make(map[string]string)
	var currentID string
	var currentValue string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if currentID != "" {
				translations[currentID] = currentValue
			}
			currentID = ""
			currentValue = ""
			continue
		}
		if strings.HasPrefix(trimmed, "msgid ") {
			if currentID != "" {
				translations[currentID] = currentValue
			}
			currentID = parsePOValue(strings.TrimSpace(strings.TrimPrefix(trimmed, "msgid ")))
			currentValue = ""
			continue
		}
		if strings.HasPrefix(trimmed, "msgstr ") {
			currentValue = parsePOValue(strings.TrimSpace(strings.TrimPrefix(trimmed, "msgstr ")))
		}
	}
	if currentID != "" {
		translations[currentID] = currentValue
	}
	return translations
}

func parsePOValue(value string) string {
	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}
	return strings.Trim(value, "\"")
}
