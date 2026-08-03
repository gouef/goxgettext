package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type message struct {
	id   string
	file string
	line int
}

type messageSet struct {
	entries map[string]*message
}

func newMessageSet() *messageSet {
	return &messageSet{entries: make(map[string]*message)}
}

func (m *messageSet) add(id, file string, line int) {
	if id == "" {
		return
	}
	if existing, ok := m.entries[id]; ok {
		if existing.file == "" {
			existing.file = file
		}
		if existing.line == 0 {
			existing.line = line
		}
		return
	}
	m.entries[id] = &message{id: id, file: file, line: line}
}

func hasExtension(extensions []string, ext string) bool {
	for _, item := range extensions {
		if strings.ToLower(item) == ext {
			return true
		}
	}
	return false
}

func collectMessages(root string) ([]message, error) {
	return collectMessagesWithConfig(root, newExtractorConfig())
}

func collectMessagesWithConfig(root string, cfg extractorConfig) ([]message, error) {
	messages := newMessageSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, ignored := range cfg.ignoreDirs {
				if strings.EqualFold(info.Name(), ignored) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !hasExtension(cfg.exts, ext) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		switch ext {
		case ".go":
			for _, msg := range extractGoSourceWithConfig(string(content), cfg) {
				messages.add(msg, path, 0)
			}
		case ".gohtml", ".html":
			for _, msg := range extractGoHTMLTemplate(string(content)) {
				messages.add(msg, path, 0)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	entries := make([]message, 0, len(messages.entries))
	for _, entry := range messages.entries {
		entries = append(entries, *entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].id < entries[j].id
	})
	return entries, nil
}

func buildPOT(entries []message) string {
	var b strings.Builder
	for _, entry := range entries {
		if entry.id == "" {
			continue
		}
		if entry.file != "" {
			fmt.Fprintf(&b, "#: %s", entry.file)
			if entry.line > 0 {
				fmt.Fprintf(&b, ":%d", entry.line)
			}
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "msgid %q\nmsgstr \"\"\n\n", entry.id)
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func buildPO(entries []message) string {
	var b strings.Builder
	b.WriteString("msgid \"\"\n")
	b.WriteString("msgstr \"\"\n")
	b.WriteString("\"Content-Type: text/plain; charset=UTF-8\\n\"\n")
	b.WriteString("\"Language: en\\n\"\n\n")
	for _, entry := range entries {
		if entry.id == "" {
			continue
		}
		fmt.Fprintf(&b, "msgid %q\nmsgstr %q\n\n", entry.id, entry.id)
	}
	return strings.TrimSpace(b.String()) + "\n"
}
